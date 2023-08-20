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
    ).resolves.not.toThrowError();
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
    ).rejects.toThrow(verifier.ValidationError);
  });

  it("fails to open non-existant file", async () => {
    const tmpFilePath = path.join(os.tmpdir(), "some-file-doesnt-exist.txt");

    await expect(
      verifier.validateFileDigest(
        tmpFilePath,
        "5aa03f96c77536579166fba147929626cc3a97960e994057a9d80271a736d10f",
      ),
    ).rejects.toThrowError();
  });
});
