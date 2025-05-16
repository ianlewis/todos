#!/usr/bin/env node

// Copyright 2025 Ian Lewis
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

function getBinaryPath() {
  // Lookup table for all platforms and binary distribution packages
  const BINARY_DISTRIBUTION_PACKAGES = {
    "linux-x64": "todos-linux-amd64",
    "linux-arm64": "todos-linux-arm64",
    "darwin-x64": "todos-darwin-amd64",
    "darwin-arm64": "todos-darwin-arm64",
    "win32-x64": "todos-windows-amd64",
    "win32-arm64": "todos-windows-arm64",
  };

  // Determine package name for this platform
  const platformSpecificPackageName =
    BINARY_DISTRIBUTION_PACKAGES[`${process.platform}-${process.arch}`];

  // Windows binaries end with .exe so we need to special case them.
  const binaryName =
    process.platform === "win32"
      ? `${platformSpecificPackageName}.exe`
      : platformSpecificPackageName;

  // Resolving will fail if the optionalDependency was not installed
  return require.resolve(`${platformSpecificPackageName}/${binaryName}`);
}

require("child_process")
  .spawn(getBinaryPath(), process.argv.slice(2), {
    stdio: "inherit",
  })
  .on("error", (err) => {
    // eslint-disable-next-line no-console
    console.error(err);
    process.exit(1);
  })
  .on("exit", (code) => process.exit(code));
