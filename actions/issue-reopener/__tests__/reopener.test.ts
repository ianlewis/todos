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

import * as fs from "fs";
import * as os from "os";
import * as path from "path";

// NOTE: must use require for mock to work.
const exec = require("@actions/exec");

const verifier = require("../src/verifier");

import * as reopener from "../src/reopener";

jest.mock("@actions/exec");
jest.mock("../src/verifier");

describe("runIssueReopener", () => {
  const env = process.env;

  beforeEach(() => {
    jest.resetModules();
    process.env = { ...env };
    process.env.RUNNER_TEMP = os.tmpdir();
    process.env.GITHUB_REPOSITORY = "github.com/owner/repo";
    process.env.GITHUB_SHA = "deadbeef";
  });

  afterEach(() => {
    process.env = env;
  });

  it("runs issue reopener", async () => {
    verifier.downloadAndVerifySLSA.mockClear();
    exec.getExecOutput.mockClear();
    const tmpDir = fs.mkdtempSync(path.join(os.tmpdir(), "runIssueReopener_"));
    const reopenerPath = path.join(tmpDir, "github-issue-reopener");
    fs.writeFileSync(reopenerPath, "");

    verifier.downloadAndVerifySLSA.mockResolvedValueOnce(reopenerPath);

    // issue-reopener
    exec.getExecOutput.mockResolvedValueOnce({
      exitCode: 0,
      stdout: "",
      stderr: "",
    });

    const workspacePath = "/home/user";
    const githubToken = "abcdef";

    await expect(
      reopener.runIssueReopener(workspacePath, githubToken, false),
    ).resolves.not.toThrow();

    expect(exec.getExecOutput).toBeCalledWith(
      reopenerPath,
      [
        `--repo=${process.env.GITHUB_REPOSITORY}`,
        `--sha=${process.env.GITHUB_SHA}`,
        workspacePath,
      ],
      {
        env: {
          GH_TOKEN: githubToken,
        },
        ignoreReturnCode: true,
      },
    );
  });

  it("runs issue reopener dry run", async () => {
    verifier.downloadAndVerifySLSA.mockClear();
    exec.getExecOutput.mockClear();
    const tmpDir = fs.mkdtempSync(path.join(os.tmpdir(), "runIssueReopener_"));
    const reopenerPath = path.join(tmpDir, "github-issue-reopener");
    fs.writeFileSync(reopenerPath, "");

    verifier.downloadAndVerifySLSA.mockResolvedValueOnce(reopenerPath);

    // issue-reopener
    exec.getExecOutput.mockResolvedValueOnce({
      exitCode: 0,
      stdout: "",
      stderr: "",
    });

    const workspacePath = "/home/user";
    const githubToken = "abcdef";

    await expect(
      // NOTE: dryRun is true.
      reopener.runIssueReopener(workspacePath, githubToken, true),
    ).resolves.not.toThrow();

    expect(exec.getExecOutput).toBeCalledWith(
      reopenerPath,
      [
        `--repo=${process.env.GITHUB_REPOSITORY}`,
        `--sha=${process.env.GITHUB_SHA}`,
        `--dry-run`,
        workspacePath,
      ],
      {
        env: {
          GH_TOKEN: githubToken,
        },
        ignoreReturnCode: true,
      },
    );
  });

  it("handles issue reopener failure", async () => {
    verifier.downloadAndVerifySLSA.mockClear();
    exec.getExecOutput.mockClear();
    const tmpDir = fs.mkdtempSync(path.join(os.tmpdir(), "runIssueReopener_"));
    const reopenerPath = path.join(tmpDir, "github-issue-reopener");
    fs.writeFileSync(reopenerPath, "");

    verifier.downloadAndVerifySLSA.mockResolvedValueOnce(reopenerPath);

    // issue-reopener
    exec.getExecOutput.mockResolvedValueOnce({
      // NOTE: exit code is returned.
      exitCode: 1,
      stdout: "",
      stderr: "",
    });

    const workspacePath = "/home/user";
    const githubToken = "abcdef";

    await expect(
      reopener.runIssueReopener(workspacePath, githubToken, false),
    ).rejects.toThrow(reopener.ReopenError);

    expect(exec.getExecOutput).toBeCalledWith(
      reopenerPath,
      [
        `--repo=${process.env.GITHUB_REPOSITORY}`,
        `--sha=${process.env.GITHUB_SHA}`,
        workspacePath,
      ],
      {
        env: {
          GH_TOKEN: githubToken,
        },
        ignoreReturnCode: true,
      },
    );
  });

  it("handles verification failure", async () => {
    verifier.downloadAndVerifySLSA.mockClear();
    exec.getExecOutput.mockClear();
    const tmpDir = fs.mkdtempSync(path.join(os.tmpdir(), "runIssueReopener_"));
    const reopenerPath = path.join(tmpDir, "github-issue-reopener");
    fs.writeFileSync(reopenerPath, "");

    // NOTE: verification error.
    verifier.downloadAndVerifySLSA.mockRejectedValueOnce(
      new verifier.VerificationError("error"),
    );

    // issue-reopener
    exec.getExecOutput.mockResolvedValueOnce({
      exitCode: 0,
      stdout: "",
      stderr: "",
    });

    const workspacePath = "/home/user";
    const githubToken = "abcdef";

    await expect(
      reopener.runIssueReopener(workspacePath, githubToken, false),
    ).rejects.toBeInstanceOf(verifier.VerificationError);

    expect(exec.getExecOutput).not.toHaveBeenCalled();
  });
});
