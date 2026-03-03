"use strict";

// Postinstall script: validates that the platform binary was installed correctly.
// Runs after `npm install @kosli-dev/cli`.

const { execFileSync } = require("child_process");
const path = require("path");
const fs = require("fs");

const SUPPORTED = {
  linux: { x64: true, arm64: true },
  darwin: { x64: true, arm64: true },
  win32: { x64: true, arm64: true },
};

const platform = process.platform;
const arch = process.arch;

if (!SUPPORTED[platform] || !SUPPORTED[platform][arch]) {
  // Not a supported platform — exit cleanly so npm install doesn't fail.
  process.exit(0);
}

const packageName = `@kosli-dev/cli-${platform}-${arch}`;

let binaryPath;
try {
  const packageDir = path.dirname(
    require.resolve(`${packageName}/package.json`)
  );
  const binaryName = platform === "win32" ? "kosli.exe" : "kosli";
  binaryPath = path.join(packageDir, "bin", binaryName);
} catch (e) {
  // Optional dependency was skipped (e.g. --no-optional). Warn but don't fail.
  process.stderr.write(
    `[kosli] Warning: ${packageName} is not installed.\n` +
    `[kosli] The kosli binary will not be available.\n` +
    `[kosli] Re-run without --no-optional to fix this.\n`
  );
  process.exit(0);
}

if (!fs.existsSync(binaryPath)) {
  process.stderr.write(`[kosli] Warning: binary not found at ${binaryPath}\n`);
  process.exit(0);
}

try {
  execFileSync(binaryPath, ["version"], { stdio: "ignore" });
} catch (e) {
  process.stderr.write(
    `[kosli] Warning: binary validation failed: ${e.message}\n`
  );
}
