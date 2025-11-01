#!/usr/bin/env node

/**
 * Post-install script for @chadneal/gimage-mcp
 *
 * This script downloads the appropriate gimage binary for the user's platform
 * from the GitHub releases page.
 */

const https = require('https');
const fs = require('fs');
const path = require('path');
const { execSync } = require('child_process');
const packageJson = require('../package.json');

// Configuration
const GITHUB_REPO = 'chadneal/gimage';
const VERSION = packageJson.version;
const BIN_DIR = path.join(__dirname, '..', 'bin');

// Platform detection
function getPlatform() {
  const platform = process.platform;
  const arch = process.arch;

  // Map Node.js platform names to release names
  const platformMap = {
    'darwin': 'Darwin',
    'linux': 'Linux',
    'win32': 'Windows'
  };

  const archMap = {
    'x64': 'x86_64',
    'arm64': 'arm64'
  };

  const mappedPlatform = platformMap[platform];
  const mappedArch = archMap[arch];

  if (!mappedPlatform || !mappedArch) {
    throw new Error(`Unsupported platform: ${platform}-${arch}`);
  }

  return {
    platform: mappedPlatform,
    arch: mappedArch,
    isWindows: platform === 'win32'
  };
}

// Construct download URL
function getDownloadUrl(platform, arch, isWindows) {
  const ext = isWindows ? 'zip' : 'tar.gz';
  const filename = `gimage_${VERSION}_${platform}_${arch}.${ext}`;
  return `https://github.com/${GITHUB_REPO}/releases/download/v${VERSION}/${filename}`;
}

// Download file
function downloadFile(url, dest) {
  return new Promise((resolve, reject) => {
    console.log(`Downloading gimage v${VERSION}...`);
    console.log(`URL: ${url}`);

    const file = fs.createWriteStream(dest);

    https.get(url, (response) => {
      if (response.statusCode === 302 || response.statusCode === 301) {
        // Follow redirect
        return downloadFile(response.headers.location, dest)
          .then(resolve)
          .catch(reject);
      }

      if (response.statusCode !== 200) {
        reject(new Error(`Download failed with status ${response.statusCode}`));
        return;
      }

      response.pipe(file);

      file.on('finish', () => {
        file.close();
        resolve();
      });
    }).on('error', (err) => {
      fs.unlink(dest, () => {}); // Clean up
      reject(err);
    });
  });
}

// Extract archive
function extractArchive(archivePath, isWindows) {
  console.log('Extracting archive...');

  try {
    if (isWindows) {
      // Windows: use PowerShell to extract zip
      execSync(`powershell -command "Expand-Archive -Path '${archivePath}' -DestinationPath '${BIN_DIR}' -Force"`, {
        stdio: 'inherit'
      });
    } else {
      // Unix: use tar
      execSync(`tar -xzf "${archivePath}" -C "${BIN_DIR}"`, {
        stdio: 'inherit'
      });
    }

    // Make binary executable (Unix only)
    if (!isWindows) {
      const binaryPath = path.join(BIN_DIR, 'gimage');
      fs.chmodSync(binaryPath, '755');
    }

    // Clean up archive
    fs.unlinkSync(archivePath);
  } catch (error) {
    throw new Error(`Failed to extract archive: ${error.message}`);
  }
}

// Main installation function
async function install() {
  try {
    console.log('Installing gimage MCP server...');

    // Detect platform
    const { platform, arch, isWindows } = getPlatform();
    console.log(`Platform: ${platform} ${arch}`);

    // Create bin directory
    if (!fs.existsSync(BIN_DIR)) {
      fs.mkdirSync(BIN_DIR, { recursive: true });
    }

    // Construct download URL
    const url = getDownloadUrl(platform, arch, isWindows);
    const ext = isWindows ? 'zip' : 'tar.gz';
    const archivePath = path.join(BIN_DIR, `gimage.${ext}`);

    // Download binary
    await downloadFile(url, archivePath);

    // Extract archive
    extractArchive(archivePath, isWindows);

    console.log('✓ gimage MCP server installed successfully!');
    console.log('');
    console.log('To configure for Claude Desktop, add to your config:');
    console.log('');
    console.log('{');
    console.log('  "mcpServers": {');
    console.log('    "gimage": {');
    console.log('      "command": "gimage-mcp",');
    console.log('      "args": ["serve"]');
    console.log('    }');
    console.log('  }');
    console.log('}');
    console.log('');
    console.log('Set up API credentials with: gimage auth gemini');

  } catch (error) {
    console.error('Installation failed:', error.message);
    console.error('');
    console.error('You can manually download from:');
    console.error(`https://github.com/${GITHUB_REPO}/releases/tag/v${VERSION}`);
    process.exit(1);
  }
}

// Run installation
if (require.main === module) {
  install();
}

module.exports = { install };
