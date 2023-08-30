/** @type {import('ts-jest').JestConfigWithTsJest} */

// Mock out stdout so that GitHub commands are ignored.
const processStdoutWrite = process.stdout.write.bind(process.stdout);
process.stdout.write = (str, encoding, cb) => {
  // Note: this will soon change to ::
  if (!str.match(/^::/)) {
    return processStdoutWrite(str, encoding, cb);
  }
};

module.exports = {
  preset: "ts-jest",
  resetMocks: true,
  testEnvironment: "node",
  collectCoverage: true,
  coverageReporters: ["text", "cobertura"],
};
