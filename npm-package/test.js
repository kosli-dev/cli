#!/usr/bin/env node

/**
 * Simple test to verify the Kosli CLI binary is installed correctly
 */

const { execSync } = require('child_process');
const path = require('path');
const fs = require('fs');

const platform = process.platform;
const binaryName = platform === 'win32' ? 'kosli.exe' : 'kosli';
const binaryPath = path.join(__dirname, 'bin', binaryName);

console.log('Testing Kosli CLI installation...');

// Check if binary exists
if (!fs.existsSync(binaryPath)) {
  console.error('✗ Binary not found at:', binaryPath);
  console.error('Installation may have failed.');
  process.exit(1);
}

console.log('✓ Binary found at:', binaryPath);

// Try to execute version command
try {
  const output = execSync(`"${binaryPath}" version`, { encoding: 'utf8' });
  console.log('✓ Binary is executable');
  console.log('\nVersion output:');
  console.log(output);
  console.log('✓ All tests passed!');
  process.exit(0);
} catch (error) {
  console.error('✗ Failed to execute binary:', error.message);
  process.exit(1);
}
