#!/usr/bin/env node

/**
 * Main entry point for the Kosli CLI npm package
 */

const { spawn } = require('child_process');
const path = require('path');

// Determine the binary name based on platform
const platform = process.platform;
const binaryName = platform === 'win32' ? 'kosli.exe' : 'kosli';
const binaryPath = path.join(__dirname, 'bin', binaryName);

// Forward all arguments to the Kosli binary
const args = process.argv.slice(2);
const kosli = spawn(binaryPath, args, { stdio: 'inherit' });

kosli.on('exit', (code) => {
  process.exit(code);
});

kosli.on('error', (err) => {
  console.error('Failed to execute Kosli CLI:', err.message);
  process.exit(1);
});
