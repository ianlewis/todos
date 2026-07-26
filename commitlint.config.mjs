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

export default {
  extends: ["@commitlint/config-conventional"],

  rules: {
    // Enforce including a scope in the commit message. This is normally
    // optional in the conventional commit specification, but we want to require
    // it for better commit message clarity.
    // https://sumnerevans.com/posts/software-engineering/stop-using-conventional-commits/
    "scope-empty": [2, "never"],
  },

  ignores: [
    // Ignore the 'Initial plan' commits created by the GitHub Copilot agent.
    // Currently the agent ignores any repository instructions when creating
    // this commit and so it can't be changed.
    // https://github.com/orgs/community/discussions/178992
    (commit) => /^[Ii]nitial plan/.test(commit),
  ],
};
