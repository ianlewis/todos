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

import * as core from "@actions/core";
import * as exec from "@actions/exec";
import * as io from "@actions/io";
import * as tc from "@actions/tool-cache";

const REOPENER_VERSION = "v0.3.0";
const SLSA_VERIFIER_VERSION = "v2.3.0";
const SLSA_VERIFIER_SHA256SUM =
  "ea687149d658efecda64d69da999efb84bb695a3212f29548d4897994027172d";

class ValidationError extends Error {
  constructor(path: string, wantDigest: string, gotDigest: string) {
    super(
      `validation error for file ${path}: expected "${wantDigest}", got "${gotDigest}"`,
    );

    // Set the prototype explicitly.
    Object.setPrototypeOf(this, ValidationError.prototype);
  }
}

// validateFileDigest validates a sha256 hex digest of the given file's contents
// against the expected digest. If a validation error occurs a ValidationError
// is thrown.
export async function validateFileDigest(
  filePath: string,
  expectedDigest: string,
): Promise<void> {
  if (!(await fs.exists(filePath))) {
    throw new Error(`file not found: ${filePath}`);
  }
  const untrustedContents = await fs.readFile(filePath);
  const computedDigest = crypto
    .createHash("sha256")
    .update(untrustedFile)
    .digest("hex");
  if (computedDigest !== expectedDigest) {
    throw new ValidationError(filePath, expectedSha256Hash, computedDigest);
  }
}

// downloadSLSAVerifier downloads the slsa-verifier binary, verifies it against
// the expected sha256 hex digest, and returns the path to the binary.
export async function downloadSLSAVerifier(
  path: string,
  version: string,
  digest: string,
): Promise<string> {
  // Download the slsa-verifier binary.
  const verifierPath = await tc.downloadTool(
    `https://github.com/slsa-framework/slsa-verifier/releases/download/${version}/slsa-verifier-linux-amd64`,
    `${path}/slsa-verifier`,
  );

  // Validate the checksum.
  await validateFileDigest(verifierPath, digest);

  return verifierPath;
}

// downloadAndVerify downloads the `github-issue-reopener` binary and verifies
// the associated SLSA provenance.
export async function downloadAndVerify(
  path: string,
  version: string,
  slsaVerifierVersion: string,
  slsaVerifierDigest: string,
): Promise<string> {
  const verifierPath = await downloadSLSAVerifier(
    path,
    slsaVerifierVersion,
    slsaVerifierDigest,
  );

  // Download the github-issue-reopener binary.
  const reopenerPath = await tc.downloadTool(
    `https://github.com/ianlewis/todos/releases/download/${version}/github-issue-reopener-linux-amd64`,
    `${path}/github-issue-reopener`,
  );

  const provenancePath = await tc.downloadTool(
    `https://github.com/ianlewis/todos/releases/download/${version}/github-issue-reopener-linux-amd64.intoto.jsonl`,
    `${path}/github-issue-reopener.intoto.jsonl`,
  );

  const { exitCode, stdout, stderr } = await exec.getExecOutput(
    `${verifierPath}`,
    [
      "verify-artifact",
      reopenerPath,
      "--provenance-path",
      provenancePath,
      "--source-uri",
      "github.com/ianlewis/todos",
      "--source-tag",
      version,
    ],
  );
  if (exitCode !== 0) {
    throw new Error(
      `unable to verify binary provenance.\nstdout: ${stdout}; stderr: ${stderr}`,
    );
  }

  return reopenerPath;
}

// cleanup deletes the temporary files.
async function cleanup(path: string): Promise<void> {
  await io.rmRF(`${path}`);
}

function run(): void {
  const path = core.getInput("path", { required: true });
  const token = core.getInput("token", { required: true });
  const tmpDir = fs.mkdtempSync(
    path.join(os.tmpdir(), "github-issue-reopener_"),
  );

  try {
    const reopenerPath = downloadAndVerify(
      tmpDir,
      REOPENER_VERSION,
      SLSA_VERIFIER_VERSION,
      SLSA_VERIFIER_SHA256SUM,
    );
  } catch (error: unknown) {
    const errMsg = error instanceof Error ? error.message : String(error);
    core.setFailed(`failed to download github-issue-reopener: ${errMsg}`);
    await cleanup(tmpDir);
    return;
  }

  // Run the github-issue-reopener.
  const { exitCode, stdout, stderr } = await exec.getExecOutput(
    `${reopenerPath}`,
    [
      `--repo=${process.env.GITHUB_REPOSITORY}`,
      `--sha=${process.env.GITHUB_SHA}`,
      `${path}`,
    ],
  );
  if (exitCode !== 0) {
    core.setFailed(
      `failed executing github-issue-reopener: exit code ${exitCode}\nstdout: ${stdout}; stderr: ${stderr}`,
    );
    await cleanup(tmpDir);
    return;
  }

  await cleanup(tmpDir);
}

run();
