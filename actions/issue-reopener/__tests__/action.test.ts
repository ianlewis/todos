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

const reopener = require("../src/reopener");

import * as action from "../src/action";

jest.mock("../src/reopener");

describe("runAction", () => {
  const env = process.env;

  beforeEach(() => {
    jest.resetModules();
    process.env = { ...env };
  });

  afterEach(() => {
    process.env = env;
  });

  it("calls runIssueReopener", async () => {
    reopener.runIssueReopener.mockClear();
    reopener.runIssueReopener.mockResolvedValueOnce(undefined);

    const workspacePath = "/home/user";
    const githubToken = "deadbeef";
    const dryRun = false;

    process.env.INPUT_PATH = workspacePath;
    process.env.INPUT_TOKEN = githubToken;
    process.env["INPUT_DRY-RUN"] = String(dryRun);

    await action.runAction();

    expect(reopener.runIssueReopener).toBeCalledWith(
      workspacePath,
      githubToken,
      dryRun,
    );

    expect(process.exitCode).toBeUndefined();
  });

  it("handles reopener failure", async () => {
    reopener.runIssueReopener.mockClear();
    reopener.runIssueReopener.mockRejectedValueOnce(new Error("foo"));

    const workspacePath = "/home/user";
    const githubToken = "deadbeef";
    const dryRun = false;

    process.env.INPUT_PATH = workspacePath;
    process.env.INPUT_TOKEN = githubToken;
    process.env["INPUT_DRY-RUN"] = String(dryRun);

    await action.runAction();

    expect(reopener.runIssueReopener).toBeCalledWith(
      workspacePath,
      githubToken,
      dryRun,
    );

    expect(process.exitCode).not.toBe(0);
  });
});
