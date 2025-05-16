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

import globals from "globals";
import github from "eslint-plugin-github";

export default [
  github.getFlatConfigs().recommended,
  {
    files: ["**/*.{js,mjs,cjs,jsx,mjsx}"],
    ignores: ["**/coverage", "**/dist", "**/linter", "**/node_modules"],
    languageOptions: {
      globals: {
        ...globals.node,
        ...globals.jest,
        Atomics: "readonly",
        SharedArrayBuffer: "readonly",
      },

      ecmaVersion: 2023,
      sourceType: "module",
    },
    rules: {
      "github/array-foreach": "error",
      "github/async-preventdefault": "warn",
      "github/no-then": "error",
      "github/no-blur": "error",
      "i18n-text/no-en": ["off"],
      "import/no-namespace": "off",
      "import/no-commonjs": "off",
      "no-unused-vars": ["error", { argsIgnorePattern: "^_" }],
    },
  },
];
