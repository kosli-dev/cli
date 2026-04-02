"use strict";

// Postinstall script: validates that the platform binary was installed correctly.
// Runs after `npm install @kosli/cli`.

const path = require("path");
const fs = require("fs");

const SUPPORTED = {
  linux: { x64: true, arm64: true, arm: true },
  darwin: { x64: true, arm64: true },
  win32: { x64: true, arm64: true },
};

const platform = process.platform;
const arch = process.arch;

if (!SUPPORTED[platform] || !SUPPORTED[platform][arch]) {
  process.stderr.write(
    `[kosli] Error: ${platform}/${arch} is not a supported platform.\n` +
    `[kosli] See https://github.com/kosli-dev/cli for supported platforms.\n` +
    `[kosli] Use --ignore-scripts to skip this check.\n`
  );
  process.exit(1);
}

const packageName = `@kosli/cli-${platform}-${arch}`;

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
  process.stderr.write(
    `[kosli] Error: binary not found at ${binaryPath}\n` +
    `[kosli] Try reinstalling: npm install -g @kosli/cli\n`
  );
  process.exit(1);
}

try {
  fs.accessSync(binaryPath, fs.constants.X_OK);
} catch (e) {
  process.stderr.write(
    `[kosli] Error: binary is not executable: ${e.message}\n` +
    `[kosli] Try reinstalling: npm install -g @kosli/cli\n`
  );
  process.exit(1);
}
