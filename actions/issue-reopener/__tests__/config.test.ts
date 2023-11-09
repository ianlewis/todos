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

import YAML from "yaml";

import * as config from "../src/config";

describe("readConfig", () => {
  it("handles non-existant config", async () => {
    let c = await config.readConfig("not-exists.yml");
    expect(c).toEqual({});
  });

  it("handles empty config", async () => {
    const tmpDir = fs.mkdtempSync(path.join(os.tmpdir(), "readConfig"));
    const configPath = path.join(tmpDir, "empty.yml");
    fs.writeFileSync(configPath, "");

    let c = await config.readConfig(configPath);

    expect(c).toEqual({});
  });

  it("handles basic config", async () => {
    const tmpDir = fs.mkdtempSync(path.join(os.tmpdir(), "readConfig"));
    const configPath = path.join(tmpDir, "todos.yml");
    fs.writeFileSync(
      configPath,
      'vanityURLs:\n  - "(https?://)?golang.org/issues/(?<id>[0-9]+)"',
    );

    let c = await config.readConfig(configPath);

    expect(c).toEqual({
      vanityURLs: ["(https?://)?golang.org/issues/(?<id>[0-9]+)"],
    });
  });

  it("handles invalid config", async () => {
    const tmpDir = fs.mkdtempSync(path.join(os.tmpdir(), "readConfig"));
    const configPath = path.join(tmpDir, "todos.yml");
    fs.writeFileSync(configPath, "vanityURLs:\n- ,");

    expect(() => config.readConfig(configPath)).rejects.toThrow(
      YAML.YAMLParseError,
    );
  });
});
