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
const tc = require("@actions/tool-cache");

import * as verifier from "../src/verifier";

jest.mock("@actions/exec");
jest.mock("@actions/tool-cache");

describe("validateFileDigest", () => {
  it("validates a file digest", async () => {
    const tmpDir = fs.mkdtempSync(
      path.join(os.tmpdir(), "validateFileDigest_"),
    );
    const tmpFilePath = path.join(tmpDir, "test.txt");
    fs.writeFileSync(tmpFilePath, "some data"); // sha256:1307990e6ba5ca145eb35e99182a9bec46531bc54ddf656a602c780fa0240dee

    await expect(
      verifier.validateFileDigest(
        tmpFilePath,
        "1307990e6ba5ca145eb35e99182a9bec46531bc54ddf656a602c780fa0240dee",
      ),
    ).resolves.not.toThrow();
  });

  it("fails validation", async () => {
    const tmpDir = fs.mkdtempSync(
      path.join(os.tmpdir(), "validateFileDigest_"),
    );
    const tmpFilePath = path.join(tmpDir, "test.txt");
    fs.writeFileSync(tmpFilePath, "some data"); // sha256:1307990e6ba5ca145eb35e99182a9bec46531bc54ddf656a602c780fa0240dee

    await expect(
      verifier.validateFileDigest(
        tmpFilePath,
        "5aa03f96c77536579166fba147929626cc3a97960e994057a9d80271a736d10f",
      ),
    ).rejects.toThrow(verifier.DigestValidationError);
  });

  it("fails to open non-existant file", async () => {
    const tmpFilePath = path.join(os.tmpdir(), "some-file-doesnt-exist.txt");

    await expect(
      verifier.validateFileDigest(
        tmpFilePath,
        "5aa03f96c77536579166fba147929626cc3a97960e994057a9d80271a736d10f",
      ),
    ).rejects.toThrow(verifier.FileError);
  });
});

describe("downloadSLSAVerifier", () => {
  const env = process.env;

  beforeEach(() => {
    jest.resetModules();
    process.env = { ...env };
    process.env.RUNNER_TEMP = os.tmpdir();
  });

  afterEach(() => {
    process.env = env;
  });

  it("downloads the verifier", async () => {
    tc.downloadTool.mockClear();

    const tmpDir = fs.mkdtempSync(
      path.join(os.tmpdir(), "downloadSLSAVerifier_"),
    );
    const tmpFilePath = path.join(tmpDir, "slsa-verifier");
    fs.writeFileSync(tmpFilePath, "some data"); // sha256:1307990e6ba5ca145eb35e99182a9bec46531bc54ddf656a602c780fa0240dee

    tc.downloadTool.mockResolvedValueOnce(tmpFilePath);

    const verifierPath = await verifier.downloadSLSAVerifier(
      "v2.3.0",
      "1307990e6ba5ca145eb35e99182a9bec46531bc54ddf656a602c780fa0240dee",
    );

    // Check that the file exists and is executable.
    expect(() => fs.accessSync(verifierPath, fs.constants.X_OK)).not.toThrow(
      Error,
    );
  });

  it("fails validation", async () => {
    tc.downloadTool.mockClear();

    const tmpDir = fs.mkdtempSync(
      path.join(os.tmpdir(), "downloadSLSAVerifier_"),
    );
    const tmpFilePath = path.join(tmpDir, "slsa-verifier");
    fs.writeFileSync(tmpFilePath, "some data"); // sha256:1307990e6ba5ca145eb35e99182a9bec46531bc54ddf656a602c780fa0240dee

    tc.downloadTool.mockResolvedValueOnce(tmpFilePath);

    await expect(
      verifier.downloadSLSAVerifier(
        "v2.3.0",
        "5aa03f96c77536579166fba147929626cc3a97960e994057a9d80271a736d10f",
      ),
    ).rejects.toThrow(verifier.DigestValidationError);
  });

  it("fails with http error", async () => {
    tc.downloadTool.mockClear();
    tc.downloadTool.mockRejectedValueOnce(new tc.HTTPError("foo"));

    await expect(
      verifier.downloadSLSAVerifier(
        "v2.3.0",
        "ea687149d658efecda64d69da999efb84bb695a3212f29548d4897994027172d",
      ),
    ).rejects.toBeInstanceOf(tc.HTTPError);
  });
});

describe("downloadAndVerifySLSA", () => {
  const env = process.env;

  beforeEach(() => {
    jest.resetModules();
    process.env = { ...env };
    process.env.RUNNER_TEMP = os.tmpdir();
  });

  afterEach(() => {
    process.env = env;
  });

  it("succeeds download", async () => {
    tc.downloadTool.mockClear();
    exec.getExecOutput.mockClear();
    const tmpDir = fs.mkdtempSync(
      path.join(os.tmpdir(), "downloadAndVerifySLSA_"),
    );

    // slsa-verifier
    const verifierPath = path.join(tmpDir, "slsa-verifier");
    fs.writeFileSync(verifierPath, "some data"); // sha256:1307990e6ba5ca145eb35e99182a9bec46531bc54ddf656a602c780fa0240dee
    tc.downloadTool.mockResolvedValueOnce(verifierPath);

    // artifact
    const artifactPath = path.join(tmpDir, "artifact");
    fs.writeFileSync(artifactPath, "artifact data"); // sha256:18eec0cf867e893de3351c2efc679c3b79ceed8a9037eec96db06a50bfd718d3
    tc.downloadTool.mockResolvedValueOnce(artifactPath);

    // provenance
    const provenancePath = path.join(tmpDir, "provenance");
    fs.writeFileSync(provenancePath, "provenance data"); // sha256:48792d9931fd23e6094bdd7c1265710a03c8b9a80800acbd71c9d623140dcf95
    tc.downloadTool.mockResolvedValueOnce(provenancePath);

    exec.getExecOutput.mockResolvedValueOnce({
      exitCode: 0,
      stdout: "PASSED",
      stderr: "",
    });

    const artifactURL =
      "https://github.com/owner/repo/releases/download/v1.0.0/artifact";
    const provenanceURL =
      "https://github.com/owner/repo/releases/download/v1.0.0/artifact.intoto.jsonl";
    const artifactRepo = "github.com/owner/repo";
    const artifactTag = "v1.0.0";
    const verifierVersion = "v2.3.0";
    const verifierDigest =
      "1307990e6ba5ca145eb35e99182a9bec46531bc54ddf656a602c780fa0240dee";

    const returnPath = await verifier.downloadAndVerifySLSA(
      artifactURL,
      provenanceURL,
      artifactRepo,
      artifactTag,
      verifierVersion,
      verifierDigest,
    );

    expect(returnPath).toBe(artifactPath);

    // Check that the file exists.
    expect(() => fs.readFileSync(returnPath, { encoding: "utf-8" })).toBe(
      "artifact data",
    );
  });

  //   it("fails SLSA verification", async () => {
  //     await expect(
  //       verifier.downloadAndVerifySLSA(
  //         "https://github.com/slsa-framework/slsa-github-generator/releases/download/v1.8.0/slsa-generator-generic-linux-amd64",
  //         "https://github.com/slsa-framework/slsa-github-generator/releases/download/v1.7.0/slsa-generator-generic-linux-amd64.intoto.jsonl",
  //         "github.com/slsa-framework/slsa-github-generator",
  //         "v1.8.0",
  //         "v2.3.0",
  //         "ea687149d658efecda64d69da999efb84bb695a3212f29548d4897994027172d",
  //       ),
  //     ).rejects.toThrow(verifier.VerificationError);
  //   }, 30000);

  //   it("fails digest validation", async () => {
  //     await expect(
  //       verifier.downloadAndVerifySLSA(
  //         "https://github.com/slsa-framework/slsa-github-generator/releases/download/v1.8.0/slsa-generator-generic-linux-amd64",
  //         "https://github.com/slsa-framework/slsa-github-generator/releases/download/v1.8.0/slsa-generator-generic-linux-amd64.intoto.jsonl",
  //         "github.com/slsa-framework/slsa-github-generator",
  //         "v1.8.0",
  //         "v2.3.0",
  //         "5aa03f96c77536579166fba147929626cc3a97960e994057a9d80271a736d10f",
  //       ),
  //     ).rejects.toThrow(verifier.DigestValidationError);
  //   }, 30000);

  //   it("fails artifact download", async () => {
  //     await expect(
  //       verifier.downloadAndVerifySLSA(
  //         "https://github.com/slsa-framework/slsa-github-generator/releases/download/v1.8.0/bad-artifact",
  //         "https://github.com/slsa-framework/slsa-github-generator/releases/download/v1.8.0/slsa-generator-generic-linux-amd64.intoto.jsonl",
  //         "github.com/slsa-framework/slsa-github-generator",
  //         "v1.8.0",
  //         "v2.3.0",
  //         "ea687149d658efecda64d69da999efb84bb695a3212f29548d4897994027172d",
  //       ),
  //     ).rejects.toThrow(tc.HTTPError);
  //   }, 30000);

  //   it("fails provenance download", async () => {
  //     await expect(
  //       verifier.downloadAndVerifySLSA(
  //         "https://github.com/slsa-framework/slsa-github-generator/releases/download/v1.8.0/slsa-generator-generic-linux-amd64",
  //         "https://github.com/slsa-framework/slsa-github-generator/releases/download/v1.8.0/bad-provenance.intoto.jsonl",
  //         "github.com/slsa-framework/slsa-github-generator",
  //         "v1.8.0",
  //         "v2.3.0",
  //         "ea687149d658efecda64d69da999efb84bb695a3212f29548d4897994027172d",
  //       ),
  //     ).rejects.toThrow(tc.HTTPError);
  //   }, 30000);

  //   it("fails verifier download", async () => {
  //     await expect(
  //       verifier.downloadAndVerifySLSA(
  //         "https://github.com/slsa-framework/slsa-github-generator/releases/download/v1.8.0/slsa-generator-generic-linux-amd64",
  //         "https://github.com/slsa-framework/slsa-github-generator/releases/download/v1.8.0/slsa-generator-generic-linux-amd64.intoto.jsonl",
  //         "github.com/slsa-framework/slsa-github-generator",
  //         "v1.8.0",
  //         "v2.3.99",
  //         "ea687149d658efecda64d69da999efb84bb695a3212f29548d4897994027172d",
  //       ),
  //     ).rejects.toThrow(tc.HTTPError);
  //   }, 30000);
});
