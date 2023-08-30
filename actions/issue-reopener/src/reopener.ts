// Copyright 2023 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

import * as fs from "fs/promises";
import * as path from "path";

import * as core from "@actions/core";
import * as exec from "@actions/exec";
import * as github from "@actions/github";

import * as verifier from "./verifier";

const TODOS_VERSION = "v0.4.0";
const SLSA_VERIFIER_VERSION = "v2.3.0";
const SLSA_VERIFIER_SHA256SUM =
  "ea687149d658efecda64d69da999efb84bb695a3212f29548d4897994027172d";

export class ReopenError extends Error {
  constructor(message: string) {
    super(message);

    // Set the prototype explicitly.
    Object.setPrototypeOf(this, ReopenError.prototype);
  }
}

// TODORef is a reference to a TODO comment.
export class TODORef {
  path = "";
  type = "";
  text = "";
  label = "";
  message = "";
  line = 0;
  comment_line = 0;
}

// TODOIssue is a GitHub issue referenced by one or more TODOs.
export class TODOIssue {
  issueID: number;
  todos: TODORef[] = [];

  constructor(issueID: number) {
    this.issueID = issueID;
  }
}

const labelMatch = new RegExp(
  "^s*((https?://)?github.com/(.+)/(.+)/issues/|#?)([0-9]+)s*$",
);

// reopenIssues downloads the todos CLI, runs it, and returns issues linked to TODOs.
export async function getTODOIssues(wd: string): Promise<TODOIssue[]> {
  const repo = github.context.repo;

  const todosPath = await verifier.downloadAndVerifySLSA(
    `https://github.com/ianlewis/todos/releases/download/${TODOS_VERSION}/todos-linux-amd64`,
    `https://github.com/ianlewis/todos/releases/download/${TODOS_VERSION}/todos-linux-amd64.intoto.jsonl`,
    "github.com/ianlewis/todos",
    TODOS_VERSION,
    SLSA_VERIFIER_VERSION,
    SLSA_VERIFIER_SHA256SUM,
  );
  core.debug(`Setting ${todosPath} as executable`);
  await fs.chmod(todosPath, 0o700);

  core.debug(`Running git to get repository root`);
  const { stdout: gitOut } = await exec.getExecOutput(
    "git",
    ["rev-parse", "--show-toplevel"],
    {
      cwd: wd,
    },
  );
  const repoRoot = gitOut.trim();

  core.debug(`Running todos (${todosPath})`);
  const { exitCode, stdout, stderr } = await exec.getExecOutput(
    todosPath,
    // TODO: get new relative directory to repoRoot
    ["--output=json", path.relative(repoRoot, wd)],
    {
      cwd: repoRoot,
      ignoreReturnCode: true,
    },
  );
  core.debug(`Ran todos (${todosPath})`);
  if (exitCode !== 0) {
    throw new ReopenError(`todos exited ${exitCode}: ${stderr}`);
  }

  // Parse stdout into list of TODORef grouped by issue.
  const issueMap = new Map<number, TODOIssue>();
  for (let line of stdout.split("\n")) {
    line = line.trim();
    if (line === "") {
      continue;
    }
    const ref: TODORef = JSON.parse(line);
    const match = ref.label.match(labelMatch);

    if (!match) {
      continue;
    }

    // NOTE: Skip the issue if it links to another repository.
    if (
      (match[3] || match[4]) &&
      (match[3] !== repo.owner || match[4] !== repo.repo)
    ) {
      continue;
    }

    if (match[5]) {
      const issueID = Number(match[5]);
      let issue = issueMap.get(issueID);
      if (!issue) {
        issue = new TODOIssue(issueID);
      }
      issue.todos.push(ref);
      issueMap.set(issueID, issue);
    }
  }

  return Array.from(issueMap.values());
}

// reopenIssues reopens issues linked to TODOs.
export async function reopenIssues(
  wd: string,
  issues: TODOIssue[],
  token: string,
  dryRun: boolean,
): Promise<void> {
  const octokit = github.getOctokit(token);
  const repo = github.context.repo;
  const sha = github.context.sha;

  for (const issueRef of issues) {
    if (issueRef.todos.length === 0) {
      continue;
    }

    const resp = await octokit.rest.issues.get({
      owner: repo.owner,
      repo: repo.repo,
      issue_number: issueRef.issueID,
    });
    const issue = resp.data;

    if (issue.state === "open") {
      continue;
    }

    let msgPrefix = "";
    if (dryRun) {
      msgPrefix = "[dry-run] ";
    }

    core.info(
      `${msgPrefix}Reopening https://github.com/${repo.owner}/${repo.repo}/issues/${issueRef.issueID} : ${issue.title}`,
    );

    if (dryRun) {
      continue;
    }

    // Reopen the issue.
    await octokit.rest.issues.update({
      owner: repo.owner,
      repo: repo.repo,
      issue_number: issueRef.issueID,
      state: "open",
    });

    let body = "There are TODOs referencing this issue:\n";
    for (const [i, todo] of issueRef.todos.entries()) {
      // NOTE: Get the path from the root of the repository.
      body += `${i + 1}. [${todo.path}:${todo.line}](https://github.com/${
        repo.owner
      }/${repo.repo}/blob/${sha}/${todo.path}#L${todo.line}): ${
        todo.message
      }\n`;
    }

    // Post the comment.
    await octokit.rest.issues.createComment({
      owner: repo.owner,
      repo: repo.repo,
      issue_number: issueRef.issueID,
      body,
    });
  }
}
