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

import * as verifier from "../src/verifier";

describe("validateFileDigest", () => {
  it("validates a file digest", async () => {
    const tmpDir = fs.mkdtempSync(
      path.join(os.tmpdir(), "validateFileDigest_"),
    );
    const tmpFilePath = path.join(tmpDir, "test.txt");
    fs.writeFileSync(tmpFilePath, "some data");

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
    fs.writeFileSync(tmpFilePath, "some data");

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
    const verifierPath = await verifier.downloadSLSAVerifier(
      "v2.3.0",
      "ea687149d658efecda64d69da999efb84bb695a3212f29548d4897994027172d",
    );

    // Check that the file exists and is executable.
    expect(() => fs.accessSync(verifierPath, fs.constants.X_OK)).not.toThrow(
      Error,
    );
  });

  it("fails validation", async () => {
    await expect(
      verifier.downloadSLSAVerifier(
        "v2.3.0",
        "5aa03f96c77536579166fba147929626cc3a97960e994057a9d80271a736d10f",
      ),
    ).rejects.toThrow(verifier.DigestValidationError);
  });

  it("fails with unknown version", async () => {
    await expect(
      verifier.downloadSLSAVerifier(
        "v2.3.99",
        "ea687149d658efecda64d69da999efb84bb695a3212f29548d4897994027172d",
      ),
    ).rejects.toThrow(verifier.DownloadError);
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

  it("succeeds verification", async () => {
    const artifactPath = await verifier.downloadAndVerifySLSA(
      "https://github.com/slsa-framework/slsa-github-generator/releases/download/v1.8.0/slsa-generator-generic-linux-amd64",
      "https://github.com/slsa-framework/slsa-github-generator/releases/download/v1.8.0/slsa-generator-generic-linux-amd64.intoto.jsonl",
      "github.com/slsa-framework/slsa-github-generator",
      "v1.8.0",
      "v2.3.0",
      "ea687149d658efecda64d69da999efb84bb695a3212f29548d4897994027172d",
    );

    // Check that the file exists.
    expect(() => fs.accessSync(artifactPath)).not.toThrow();
  }, 30000);

  it("fails SLSA verification", async () => {
    await expect(
      verifier.downloadAndVerifySLSA(
        "https://github.com/slsa-framework/slsa-github-generator/releases/download/v1.8.0/slsa-generator-generic-linux-amd64",
        "https://github.com/slsa-framework/slsa-github-generator/releases/download/v1.7.0/slsa-generator-generic-linux-amd64.intoto.jsonl",
        "github.com/slsa-framework/slsa-github-generator",
        "v1.8.0",
        "v2.3.0",
        "ea687149d658efecda64d69da999efb84bb695a3212f29548d4897994027172d",
      ),
    ).rejects.toThrow(verifier.VerificationError);
  }, 30000);

  it("fails digest validation", async () => {
    await expect(
      verifier.downloadAndVerifySLSA(
        "https://github.com/slsa-framework/slsa-github-generator/releases/download/v1.8.0/slsa-generator-generic-linux-amd64",
        "https://github.com/slsa-framework/slsa-github-generator/releases/download/v1.8.0/slsa-generator-generic-linux-amd64.intoto.jsonl",
        "github.com/slsa-framework/slsa-github-generator",
        "v1.8.0",
        "v2.3.0",
        "5aa03f96c77536579166fba147929626cc3a97960e994057a9d80271a736d10f",
      ),
    ).rejects.toThrow(verifier.DigestValidationError);
  }, 30000);

  it("fails artifact download", async () => {
    await expect(
      verifier.downloadAndVerifySLSA(
        "https://github.com/slsa-framework/slsa-github-generator/releases/download/v1.8.0/bad-artifact",
        "https://github.com/slsa-framework/slsa-github-generator/releases/download/v1.8.0/slsa-generator-generic-linux-amd64.intoto.jsonl",
        "github.com/slsa-framework/slsa-github-generator",
        "v1.8.0",
        "v2.3.0",
        "ea687149d658efecda64d69da999efb84bb695a3212f29548d4897994027172d",
      ),
    ).rejects.toThrow(verifier.DownloadError);
  }, 30000);

  it("fails provenance download", async () => {
    await expect(
      verifier.downloadAndVerifySLSA(
        "https://github.com/slsa-framework/slsa-github-generator/releases/download/v1.8.0/slsa-generator-generic-linux-amd64",
        "https://github.com/slsa-framework/slsa-github-generator/releases/download/v1.8.0/bad-provenance.intoto.jsonl",
        "github.com/slsa-framework/slsa-github-generator",
        "v1.8.0",
        "v2.3.0",
        "ea687149d658efecda64d69da999efb84bb695a3212f29548d4897994027172d",
      ),
    ).rejects.toThrow(verifier.DownloadError);
  }, 30000);
});
