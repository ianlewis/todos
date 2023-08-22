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

import * as core from "@actions/core";
import * as exec from "@actions/exec";

import * as verifier from "./verifier";

const TODOS_VERSION = "v0.3.0";
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
  filePath: string;
  line: number;
  message: string;
}

// TODOIssue is a GitHub issue referenced by one or more TODOs.
export class TODOIssue {
  issueRef string;
  todos Array<TODORef>;
}

// reopenIssues downloads the todos CLI, runs it, and returns issues linked to TODOs.
export async function getTODOIssues(wd: string): Promise<Array<TODOIssue>> {
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

  core.debug(`Running todos (${todosPath})`);

  const { exitCode, stdout, stderr } = await exec.getExecOutput(
    todosPath,
    [wd],
    { ignoreReturnCode: true },
  );
  core.debug(`Ran todos (${todosPath})`);
  if (exitCode !== 0) {
    throw new ReopenError(`todos exited ${exitCode}: ${stderr}`);
  }

  // TODO: Parse stdout into lint of issues.
  const labelMatch = /^\s*((https?://)?github.com/(.*)/(.*)/issues/|#?)([0-9]+)\s*$/

  for (var line of stdout.split("\n")) {
    core.debug(line);
  }
}

// reopenIssues reopens issues linked to TODOs.
export async function reopenIssues(
  issues: Array<TODOIssue>,
  token: string,
  dryRun: boolean,
): Promise<void> {
  // TODO: Create a GitHub API client.
  // TODO: loop over issues and reopen them.
}
