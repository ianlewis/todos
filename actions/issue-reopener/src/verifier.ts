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

import * as crypto from "crypto";
import * as fs from "fs/promises";

import * as core from "@actions/core";
import * as exec from "@actions/exec";
import * as tc from "@actions/tool-cache";

export class FileError extends Error {
  constructor(filePath: string, message: string) {
    super(`${filePath}: ${message}`);

    // Set the prototype explicitly.
    Object.setPrototypeOf(this, FileError.prototype);
  }
}

export class DigestValidationError extends Error {
  constructor(filePath: string, wantDigest: string, gotDigest: string) {
    super(
      `validation error for file ${filePath}: expected "${wantDigest}", got "${gotDigest}"`,
    );

    // Set the prototype explicitly.
    Object.setPrototypeOf(this, DigestValidationError.prototype);
  }
}

// validateFileDigest validates a sha256 hex digest of the given file's contents
// against the expected digest. If a validation error occurs a DigestValidationError
// is thrown.
export async function validateFileDigest(
  filePath: string,
  expectedDigest: string,
): Promise<void> {
  core.debug(`Validating file digest: ${filePath}`);

  core.debug(`Expected digest for ${filePath}: ${expectedDigest}`);

  let computedDigest: string;
  try {
    // Verify that the file exists.
    await fs.access(filePath);

    const untrustedContents = await fs.readFile(filePath);
    computedDigest = crypto
      .createHash("sha256")
      .update(untrustedContents)
      .digest("hex");
  } catch (err) {
    const message = err instanceof Error ? err.message : `${err}`;
    throw new FileError(filePath, message);
  }

  core.debug(`Computed digest for ${filePath}: ${computedDigest}`);

  if (computedDigest !== expectedDigest) {
    throw new DigestValidationError(filePath, expectedDigest, computedDigest);
  }

  core.debug(`Digest for ${filePath} validated`);
}

// downloadSLSAVerifier downloads the slsa-verifier binary, verifies it against
// the expected sha256 hex digest, and returns the path to the binary.
export async function downloadSLSAVerifier(
  version: string,
  digest: string,
): Promise<string> {
  core.debug(`Downloading slsa-verifier ${version}`);

  // Download the slsa-verifier binary.
  const verifierPath = await tc.downloadTool(
    `https://github.com/slsa-framework/slsa-verifier/releases/download/${version}/slsa-verifier-linux-amd64`,
  );

  core.debug(`Downloaded slsa-verifier to ${verifierPath}`);

  // Validate the checksum.
  await validateFileDigest(verifierPath, digest);

  core.debug(`Setting ${verifierPath} as executable`);

  await fs.chmod(verifierPath, 0o700);

  return verifierPath;
}

export class VerificationError extends Error {
  constructor(message: string) {
    super(`failed to verify binary provenance: ${message}`);

    // Set the prototype explicitly.
    Object.setPrototypeOf(this, VerificationError.prototype);
  }
}

// downloadAndVerifySLSA downloads a file and verifies the associated SLSA
// provenance.
export async function downloadAndVerifySLSA(
  url: string,
  provenanceURL: string,
  sourceURI: string,
  sourceTag: string,
  slsaVerifierVersion: string,
  slsaVerifierDigest: string,
): Promise<string> {
  const verifierPromise = await downloadSLSAVerifier(
    slsaVerifierVersion,
    slsaVerifierDigest,
  );

  core.debug(`Downloading ${url}`);
  const artifactPromise = await tc.downloadTool(url);

  core.debug(`Downloading ${provenanceURL}`);
  const provenancePath = await tc.downloadTool(provenanceURL);
  core.debug(`Downloaded ${provenanceURL} to ${provenancePath}`);

  const verifierPath = await verifierPromise;
  const artifactPath = await artifactPromise;
  core.debug(`Downloaded ${url} to ${artifactPath}`);

  core.debug(`Running slsa-verifier (${verifierPath})`);

  const { exitCode, stdout, stderr } = await exec.getExecOutput(
    verifierPath,
    [
      "verify-artifact",
      artifactPath,
      "--provenance-path",
      provenancePath,
      "--source-uri",
      sourceURI,
      "--source-tag",
      sourceTag,
    ],
    { ignoreReturnCode: true },
  );

  core.debug(`Ran slsa-verifier (${verifierPath}): ${stdout}`);
  if (exitCode !== 0) {
    throw new VerificationError(`slsa-verifier exited ${exitCode}: ${stderr}`);
  }

  return artifactPath;
}
