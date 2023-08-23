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
const github = require("@actions/github");

const verifier = require("../src/verifier");
import * as reopener from "../src/reopener";

jest.mock("@actions/exec");
jest.mock("@actions/github");
jest.mock("../src/verifier");

describe("TODORef", () => {
  it("constructs", async () => {
    const ref = new reopener.TODORef();
    expect(ref.path).toBe("");
    expect(ref.type).toBe("");
    expect(ref.text).toBe("");
    expect(ref.label).toBe("");
    expect(ref.message).toBe("");
    expect(ref.line).toBe(0);
    expect(ref.comment_line).toBe(0);
  });
});

describe("getTODOIssues", () => {
  const env = process.env;

  beforeEach(() => {
    jest.resetModules();
    process.env = { ...env };

    github.context.repo = {
      owner: "owner",
      repo: "repo",
    };
  });

  it("parses empty output", async () => {
    verifier.downloadAndVerifySLSA.mockClear();
    exec.getExecOutput.mockClear();
    const tmpDir = fs.mkdtempSync(path.join(os.tmpdir(), "getTODOIssues_"));
    const todosPath = path.join(tmpDir, "todos");
    fs.writeFileSync(todosPath, "");

    verifier.downloadAndVerifySLSA.mockResolvedValueOnce(todosPath);

    // todos
    exec.getExecOutput.mockResolvedValueOnce({
      exitCode: 0,
      stdout: "",
      stderr: "",
    });

    const workspacePath = process.env.GITHUB_WORKSPACE as string;
    await expect(reopener.getTODOIssues(workspacePath)).resolves.toHaveLength(
      0,
    );
  });

  it("skips non-match", async () => {
    verifier.downloadAndVerifySLSA.mockClear();
    exec.getExecOutput.mockClear();
    const tmpDir = fs.mkdtempSync(path.join(os.tmpdir(), "getTODOIssues_"));
    const todosPath = path.join(tmpDir, "todos");
    fs.writeFileSync(todosPath, "");

    verifier.downloadAndVerifySLSA.mockResolvedValueOnce(todosPath);

    // todos
    exec.getExecOutput.mockResolvedValueOnce({
      exitCode: 0,
      stdout: '{"label": "ianlewis"}',
      stderr: "",
    });

    const workspacePath = process.env.GITHUB_WORKSPACE as string;
    await expect(reopener.getTODOIssues(workspacePath)).resolves.toHaveLength(
      0,
    );
  });

  it("skips links to other repos", async () => {
    verifier.downloadAndVerifySLSA.mockClear();
    exec.getExecOutput.mockClear();
    const tmpDir = fs.mkdtempSync(path.join(os.tmpdir(), "getTODOIssues_"));
    const todosPath = path.join(tmpDir, "todos");
    fs.writeFileSync(todosPath, "");

    verifier.downloadAndVerifySLSA.mockResolvedValueOnce(todosPath);

    // todos
    exec.getExecOutput.mockResolvedValueOnce({
      exitCode: 0,
      stdout: '{"label": "github.com/octocat/repo/issues/123"}',
      stderr: "",
    });

    const workspacePath = process.env.GITHUB_WORKSPACE as string;
    await expect(reopener.getTODOIssues(workspacePath)).resolves.toHaveLength(
      0,
    );
  });

  it("handles malformed url", async () => {
    verifier.downloadAndVerifySLSA.mockClear();
    exec.getExecOutput.mockClear();
    const tmpDir = fs.mkdtempSync(path.join(os.tmpdir(), "getTODOIssues_"));
    const todosPath = path.join(tmpDir, "todos");
    fs.writeFileSync(todosPath, "");

    verifier.downloadAndVerifySLSA.mockResolvedValueOnce(todosPath);

    // todos
    exec.getExecOutput.mockResolvedValueOnce({
      exitCode: 0,
      stdout: '{"label": "github.com//repo/issues/123"}',
      stderr: "",
    });

    const workspacePath = process.env.GITHUB_WORKSPACE as string;
    await expect(reopener.getTODOIssues(workspacePath)).resolves.toHaveLength(
      0,
    );
  });

  it("matches issue number only", async () => {
    verifier.downloadAndVerifySLSA.mockClear();
    exec.getExecOutput.mockClear();
    const tmpDir = fs.mkdtempSync(path.join(os.tmpdir(), "getTODOIssues_"));
    const todosPath = path.join(tmpDir, "todos");
    fs.writeFileSync(todosPath, "");

    verifier.downloadAndVerifySLSA.mockResolvedValueOnce(todosPath);

    // todos
    exec.getExecOutput.mockResolvedValueOnce({
      exitCode: 0,
      stdout: '{"label": "123"}',
      stderr: "",
    });

    const workspacePath = process.env.GITHUB_WORKSPACE as string;
    let p = reopener.getTODOIssues(workspacePath);
    await expect(p).resolves.toHaveLength(1);
    let issues = await p;

    expect(issues[0].todos).toHaveLength(1);
    expect(issues[0].issueID).toBe(123);
    expect(issues[0].todos[0].label).toBe("123");
  });

  it("matches issue number with #", async () => {
    verifier.downloadAndVerifySLSA.mockClear();
    exec.getExecOutput.mockClear();
    const tmpDir = fs.mkdtempSync(path.join(os.tmpdir(), "getTODOIssues_"));
    const todosPath = path.join(tmpDir, "todos");
    fs.writeFileSync(todosPath, "");

    verifier.downloadAndVerifySLSA.mockResolvedValueOnce(todosPath);

    // todos
    exec.getExecOutput.mockResolvedValueOnce({
      exitCode: 0,
      stdout: '{"label": "#123"}',
      stderr: "",
    });

    const workspacePath = process.env.GITHUB_WORKSPACE as string;
    let p = reopener.getTODOIssues(workspacePath);
    await expect(p).resolves.toHaveLength(1);
    let issues = await p;

    expect(issues[0].todos).toHaveLength(1);
    expect(issues[0].issueID).toBe(123);
    expect(issues[0].todos[0].label).toBe("#123");
  });

  it("matches issue url", async () => {
    verifier.downloadAndVerifySLSA.mockClear();
    exec.getExecOutput.mockClear();
    const tmpDir = fs.mkdtempSync(path.join(os.tmpdir(), "getTODOIssues_"));
    const todosPath = path.join(tmpDir, "todos");
    fs.writeFileSync(todosPath, "");

    verifier.downloadAndVerifySLSA.mockResolvedValueOnce(todosPath);

    // todos
    exec.getExecOutput.mockResolvedValueOnce({
      exitCode: 0,
      stdout: '{"label": "github.com/owner/repo/issues/123"}',
      stderr: "",
    });

    const workspacePath = process.env.GITHUB_WORKSPACE as string;
    let p = reopener.getTODOIssues(workspacePath);
    await expect(p).resolves.toHaveLength(1);
    let issues = await p;

    expect(issues[0].todos).toHaveLength(1);
    expect(issues[0].issueID).toBe(123);
    expect(issues[0].todos[0].label).toBe("github.com/owner/repo/issues/123");
  });

  it("groups todo", async () => {
    verifier.downloadAndVerifySLSA.mockClear();
    exec.getExecOutput.mockClear();
    const tmpDir = fs.mkdtempSync(path.join(os.tmpdir(), "getTODOIssues_"));
    const todosPath = path.join(tmpDir, "todos");
    fs.writeFileSync(todosPath, "");

    verifier.downloadAndVerifySLSA.mockResolvedValueOnce(todosPath);

    // todos
    exec.getExecOutput.mockResolvedValueOnce({
      exitCode: 0,
      stdout: '{"label": "123"}\n{"label": "456"}\n{"label": "123"}',
      stderr: "",
    });

    const workspacePath = process.env.GITHUB_WORKSPACE as string;
    let p = reopener.getTODOIssues(workspacePath);
    await expect(p).resolves.toHaveLength(2);
    let issues = await p;

    // NOTE: This may be flaky depending on Map implementation.
    // TODO: Update test to not depend on order of Map.values()
    expect(issues[0].issueID).toBe(123);
    expect(issues[0].todos).toHaveLength(2);
    expect(issues[1].issueID).toBe(456);
    expect(issues[1].todos).toHaveLength(1);
  });

  it("handles todos error", async () => {
    verifier.downloadAndVerifySLSA.mockClear();
    exec.getExecOutput.mockClear();
    const tmpDir = fs.mkdtempSync(path.join(os.tmpdir(), "getTODOIssues_"));
    const todosPath = path.join(tmpDir, "todos");
    fs.writeFileSync(todosPath, "");

    verifier.downloadAndVerifySLSA.mockResolvedValueOnce(todosPath);

    // todos
    exec.getExecOutput.mockResolvedValueOnce({
      // NOTE: todos returns exit code
      exitCode: 1,
      stdout: "",
      stderr: "ERROR",
    });

    const workspacePath = "/home/user";

    await expect(reopener.getTODOIssues(workspacePath)).rejects.toBeInstanceOf(
      reopener.ReopenError,
    );
  });
});

describe("reopenIssues", () => {
  const env = process.env;

  beforeEach(() => {
    jest.resetModules();
    process.env = { ...env };
    process.env.GITHUB_WORKSPACE = "/home/user";
  });

  it("handles empty list", async () => {
    const issues = {
      get: jest.fn(),
      createComment: jest.fn(),
      update: jest.fn(),
    };
    github.getOctokit.mockImplementation(() => {
      return {
        rest: { issues },
      };
    });

    await expect(reopener.reopenIssues([], "", false)).resolves.toBeUndefined();

    expect(issues.get).not.toHaveBeenCalled();
    expect(issues.createComment).not.toHaveBeenCalled();
    expect(issues.update).not.toHaveBeenCalled();
  });

  it("handles empty todo refs", async () => {
    const issues = {
      get: jest.fn(),
      createComment: jest.fn(),
      update: jest.fn(),
    };
    github.getOctokit.mockImplementation(() => {
      return {
        rest: { issues },
      };
    });

    // NOTE: todoIssue.todos is empty.
    const todoIssue = new reopener.TODOIssue(123);

    await expect(
      reopener.reopenIssues([todoIssue], "", false),
    ).resolves.toBeUndefined();

    expect(issues.get).not.toHaveBeenCalled();
    expect(issues.createComment).not.toHaveBeenCalled();
    expect(issues.update).not.toHaveBeenCalled();
  });

  it("handles open issue", async () => {
    const issues = {
      get: jest.fn().mockImplementation(() => {
        return {
          data: {
            state: "open",
          },
        };
      }),
      createComment: jest.fn(),
      update: jest.fn(),
    };
    github.getOctokit.mockImplementation(() => {
      return {
        rest: { issues },
      };
    });

    const todoIssue = new reopener.TODOIssue(123);
    todoIssue.todos.push(new reopener.TODORef());

    await expect(
      reopener.reopenIssues([todoIssue], "", false),
    ).resolves.toBeUndefined();

    expect(issues.get).toHaveBeenCalledTimes(1);
    expect(issues.createComment).not.toHaveBeenCalled();
    expect(issues.update).not.toHaveBeenCalled();
  });

  it("reopens closed issue", async () => {
    const issues = {
      get: jest.fn().mockImplementation(() => {
        return {
          data: {
            state: "closed",
          },
        };
      }),
      createComment: jest.fn(),
      update: jest.fn(),
    };
    github.getOctokit.mockImplementation(() => {
      return {
        rest: { issues },
      };
    });

    const todoIssue = new reopener.TODOIssue(123);
    // NOTE: multiple TODO references.
    todoIssue.todos.push(new reopener.TODORef());
    todoIssue.todos.push(new reopener.TODORef());

    await expect(
      reopener.reopenIssues([todoIssue], "", false),
    ).resolves.toBeUndefined();

    expect(issues.get).toHaveBeenCalledTimes(1);
    expect(issues.createComment).toHaveBeenCalledTimes(1);
    expect(issues.update).toHaveBeenCalledTimes(1);
  });

  it("reopens closed issue dry-run", async () => {
    const issues = {
      get: jest.fn().mockImplementation(() => {
        return {
          data: {
            state: "closed",
          },
        };
      }),
      createComment: jest.fn(),
      update: jest.fn(),
    };
    github.getOctokit.mockImplementation(() => {
      return {
        rest: { issues },
      };
    });

    const todoIssue = new reopener.TODOIssue(123);
    // NOTE: multiple TODO references.
    todoIssue.todos.push(new reopener.TODORef());
    todoIssue.todos.push(new reopener.TODORef());

    await expect(
      // NOTE: dry-run = true
      reopener.reopenIssues([todoIssue], "", true),
    ).resolves.toBeUndefined();

    expect(issues.get).toHaveBeenCalledTimes(1);
    expect(issues.createComment).not.toHaveBeenCalled();
    expect(issues.update).not.toHaveBeenCalled();
  });

  it("reopens multiple closed issues", async () => {
    const issues = {
      get: jest.fn().mockImplementation(() => {
        return {
          data: {
            state: "closed",
          },
        };
      }),
      createComment: jest.fn(),
      update: jest.fn(),
    };
    github.getOctokit.mockImplementation(() => {
      return {
        rest: { issues },
      };
    });

    const todoIssue1 = new reopener.TODOIssue(123);
    todoIssue1.todos.push(new reopener.TODORef());
    todoIssue1.todos.push(new reopener.TODORef());

    const todoIssue2 = new reopener.TODOIssue(456);
    todoIssue2.todos.push(new reopener.TODORef());

    await expect(
      reopener.reopenIssues([todoIssue1, todoIssue2], "", false),
    ).resolves.toBeUndefined();

    expect(issues.get).toHaveBeenCalledTimes(2);
    expect(issues.createComment).toHaveBeenCalledTimes(2);
    expect(issues.update).toHaveBeenCalledTimes(2);
  });
});
