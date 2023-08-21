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

const REOPENER_VERSION = "v0.3.0";
const SLSA_VERIFIER_VERSION = "v2.3.0";
const SLSA_VERIFIER_SHA256SUM =
  "ea687149d658efecda64d69da999efb84bb695a3212f29548d4897994027172d";

// runIssueReopener downloads and runs github-issue-reopener.
export async function runIssueReopener(
  wd: string,
  token: string,
  dryRun: boolean,
): Promise<void> {
  const reopenerPath = await verifier.downloadAndVerifySLSA(
    `https://github.com/ianlewis/todos/releases/download/${REOPENER_VERSION}/github-issue-reopener-linux-amd64`,
    `https://github.com/ianlewis/todos/releases/download/${REOPENER_VERSION}/github-issue-reopener-linux-amd64.intoto.jsonl`,
    "github.com/ianlewis/todos",
    REOPENER_VERSION,
    SLSA_VERIFIER_VERSION,
    SLSA_VERIFIER_SHA256SUM,
  );

  core.debug(`Setting ${reopenerPath} as executable`);

  await fs.chmod(reopenerPath, 0o700);

  // Run the github-issue-reopener.
  core.debug(`Running github-issue-reopener (${reopenerPath})`);

  let args = [
    `--repo=${process.env.GITHUB_REPOSITORY}`,
    `--sha=${process.env.GITHUB_SHA}`,
  ];
  if (dryRun) {
    args.push("--dry-run");
  }
  args.push(wd);

  const { exitCode, stdout, stderr } = await exec.getExecOutput(
    `${reopenerPath}`,
    args,
    {
      env: {
        GH_TOKEN: token,
      },
    },
  );
  if (exitCode !== 0) {
    core.setFailed(
      `failed executing github-issue-reopener: exit code ${exitCode}\nstdout: ${stdout}; stderr: ${stderr}`,
    );
    return;
  }

  core.debug(`github-issue-reopener (${reopenerPath}) exited successfully`);
}
