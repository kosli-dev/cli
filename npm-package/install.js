#!/usr/bin/env node

/**
 * This script downloads the platform-specific Kosli CLI binary
 * during npm package installation.
 */

const fs = require('fs');
const path = require('path');
const https = require('https');
const { execSync } = require('child_process');

// Get package version
const packageJson = require('./package.json');
const version = packageJson.version;

// Determine platform and architecture
const platform = process.platform;
const arch = process.arch;

// Map Node.js platform/arch to Goreleaser naming
const platformMap = {
  darwin: 'Darwin',
  linux: 'Linux',
  win32: 'Windows'
};

const archMap = {
  x64: 'x86_64',
  arm64: 'arm64',
  arm: 'arm'
};

const mappedPlatform = platformMap[platform];
const mappedArch = archMap[arch];

if (!mappedPlatform || !mappedArch) {
  console.error(`Unsupported platform: ${platform} ${arch}`);
  process.exit(1);
}

// Construct download URL
// Format: kosli_<OS>_<ARCH>.tar.gz (for Linux/Darwin)
// Format: kosli_<OS>_<ARCH>.zip (for Windows)
const extension = platform === 'win32' ? 'zip' : 'tar.gz';
const archiveName = `kosli_${mappedPlatform}_${mappedArch}.${extension}`;
const downloadUrl = `https://github.com/kosli-dev/cli/releases/download/v${version}/${archiveName}`;

console.log(`Downloading Kosli CLI v${version} for ${platform}/${arch}...`);
console.log(`URL: ${downloadUrl}`);

const binDir = path.join(__dirname, 'bin');
const archivePath = path.join(binDir, archiveName);
const binaryName = platform === 'win32' ? 'kosli.exe' : 'kosli';
const binaryPath = path.join(binDir, binaryName);

// Check if binary already exists (e.g., from local packaging)
if (fs.existsSync(binaryPath)) {
  console.log('✓ Kosli CLI binary already present, skipping download');
  console.log(`Binary location: ${binaryPath}`);
  process.exit(0);
}

// Create bin directory
if (!fs.existsSync(binDir)) {
  fs.mkdirSync(binDir, { recursive: true });
}

/**
 * Download file from URL
 */
function downloadFile(url, dest) {
  return new Promise((resolve, reject) => {
    const file = fs.createWriteStream(dest);

    https.get(url, { headers: { 'User-Agent': 'kosli-npm-installer' } }, (response) => {
      // Follow redirects
      if (response.statusCode === 301 || response.statusCode === 302) {
        return downloadFile(response.headers.location, dest).then(resolve).catch(reject);
      }

      if (response.statusCode !== 200) {
        reject(new Error(`Failed to download: HTTP ${response.statusCode}`));
        return;
      }

      response.pipe(file);

      file.on('finish', () => {
        file.close();
        resolve();
      });
    }).on('error', (err) => {
      fs.unlink(dest, () => {}); // Delete the file on error
      reject(err);
    });

    file.on('error', (err) => {
      fs.unlink(dest, () => {}); // Delete the file on error
      reject(err);
    });
  });
}

/**
 * Extract archive
 */
function extractArchive() {
  try {
    if (platform === 'win32') {
      // Extract zip file on Windows
      // Note: This requires PowerShell or a zip utility
      execSync(`tar -xf "${archivePath}" -C "${binDir}"`, { stdio: 'inherit' });
    } else {
      // Extract tar.gz on Unix-like systems
      execSync(`tar -xzf "${archivePath}" -C "${binDir}"`, { stdio: 'inherit' });
    }

    // Make binary executable on Unix-like systems
    if (platform !== 'win32') {
      fs.chmodSync(binaryPath, '755');
    }

    // Clean up archive
    fs.unlinkSync(archivePath);

    console.log('✓ Kosli CLI installed successfully!');
    console.log(`Binary location: ${binaryPath}`);
  } catch (error) {
    console.error('Failed to extract archive:', error.message);
    process.exit(1);
  }
}

// Main installation flow
downloadFile(downloadUrl, archivePath)
  .then(() => {
    console.log('✓ Download complete');
    extractArchive();
  })
  .catch((error) => {
    console.error('Failed to download Kosli CLI:', error.message);
    console.error('\nYou can manually download the binary from:');
    console.error(`https://github.com/kosli-dev/cli/releases/tag/v${version}`);
    process.exit(1);
  });
