#!/usr/bin/env node

const https = require('https');
const fs = require('fs');
const path = require('path');
const tar = require('tar');

const GITHUB_REPO = 'chadneal/gimage';
const VERSION = process.env.npm_package_version || '0.1.0';

function getPlatformInfo() {
  const platform = process.platform;
  const arch = process.arch;

  const platformMap = {
    darwin: 'darwin',
    linux: 'linux',
    win32: 'windows'
  };

  const archMap = {
    x64: 'amd64',
    arm64: 'arm64'
  };

  const mappedPlatform = platformMap[platform];
  const mappedArch = archMap[arch];

  if (!mappedPlatform || !mappedArch) {
    throw new Error(`Unsupported platform: ${platform}-${arch}`);
  }

  return {
    platform: mappedPlatform,
    arch: mappedArch,
    ext: platform === 'win32' ? '.exe' : ''
  };
}

async function downloadBinary() {
  const { platform, arch, ext } = getPlatformInfo();
  const binaryName = `gimage${ext}`;

  // GoReleaser naming format: gimage_VERSION_Platform_arch.tar.gz or .zip for Windows
  const platformCap = platform.charAt(0).toUpperCase() + platform.slice(1);
  let archName = arch;
  if (arch === 'amd64') archName = 'x86_64';

  const extension = platform === 'windows' ? '.zip' : '.tar.gz';
  const tarballName = `gimage_${VERSION}_${platformCap}_${archName}${extension}`;
  const url = `https://github.com/${GITHUB_REPO}/releases/download/v${VERSION}/${tarballName}`;

  const binDir = path.join(__dirname, '..', 'bin');
  const binaryPath = path.join(binDir, binaryName);

  // Create bin directory if it doesn't exist
  if (!fs.existsSync(binDir)) {
    fs.mkdirSync(binDir, { recursive: true });
  }

  console.log(`Downloading gimage binary for ${platform}-${arch}...`);
  console.log(`URL: ${url}`);

  return new Promise((resolve, reject) => {
    https.get(url, (response) => {
      if (response.statusCode === 302 || response.statusCode === 301) {
        // Follow redirect
        https.get(response.headers.location, (redirectResponse) => {
          if (redirectResponse.statusCode !== 200) {
            reject(new Error(`Download failed with status ${redirectResponse.statusCode}`));
            return;
          }

          const tarPath = path.join(binDir, tarballName);
          const file = fs.createWriteStream(tarPath);

          redirectResponse.pipe(file);

          file.on('finish', async () => {
            file.close();

            try {
              // Extract tarball
              await tar.x({
                file: tarPath,
                cwd: binDir
              });

              // Remove tarball
              fs.unlinkSync(tarPath);

              // Make binary executable (Unix-like systems)
              if (platform !== 'windows') {
                fs.chmodSync(binaryPath, 0o755);
              }

              console.log('✓ gimage binary installed successfully');
              resolve();
            } catch (err) {
              reject(err);
            }
          });

          file.on('error', reject);
        }).on('error', reject);
      } else if (response.statusCode === 200) {
        const tarPath = path.join(binDir, tarballName);
        const file = fs.createWriteStream(tarPath);

        response.pipe(file);

        file.on('finish', async () => {
          file.close();

          try {
            // Extract tarball
            await tar.x({
              file: tarPath,
              cwd: binDir
            });

            // Remove tarball
            fs.unlinkSync(tarPath);

            // Make binary executable (Unix-like systems)
            if (platform !== 'windows') {
              fs.chmodSync(binaryPath, 0o755);
            }

            console.log('✓ gimage binary installed successfully');
            resolve();
          } catch (err) {
            reject(err);
          }
        });

        file.on('error', reject);
      } else {
        reject(new Error(`Download failed with status ${response.statusCode}`));
      }
    }).on('error', reject);
  });
}

async function main() {
  try {
    await downloadBinary();
    console.log('\n✓ Installation complete!');
    console.log('\nTo use with Claude Desktop, add this to your MCP configuration:');
    console.log('\nmacOS: ~/Library/Application Support/Claude/claude_desktop_config.json');
    console.log('Linux: ~/.config/Claude/claude_desktop_config.json');
    console.log('Windows: %APPDATA%\\Claude\\claude_desktop_config.json');
    console.log('\n{');
    console.log('  "mcpServers": {');
    console.log('    "gimage": {');
    console.log('      "command": "npx",');
    console.log('      "args": ["-y", "@chadneal/gimage-mcp"]');
    console.log('    }');
    console.log('  }');
    console.log('}');
    console.log('\nBefore using, configure your API keys:');
    console.log('  gimage auth gemini');
    console.log('\nFor more information: https://github.com/chadneal/gimage');
  } catch (error) {
    console.error('Installation failed:', error.message);
    console.error('\nFallback options:');
    console.error('1. Install gimage manually from: https://github.com/' + GITHUB_REPO + '/releases');
    console.error('2. Install via Homebrew: brew install gimage');
    console.error('3. Build from source: https://github.com/' + GITHUB_REPO + '#building-from-source');
    process.exit(1);
  }
}

main();
