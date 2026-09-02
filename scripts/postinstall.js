#!/usr/bin/env node
// postinstall.js — downloads the zhizai binary for the current platform.

'use strict';

const fs = require('fs');
const path = require('path');
const os = require('os');
const https = require('https');
const crypto = require('crypto');
const { spawnSync } = require('child_process');

const pkg = require('../package.json');
const VERSION = pkg.version;
const REPO = 'BoteAI/zhizai-cli';

function getPlatform() {
  const platform = os.platform();
  const arch = os.arch();

  const platformMap = { darwin: 'darwin', linux: 'linux', win32: 'windows' };
  const archMap = { x64: 'amd64', arm64: 'arm64' };

  const p = platformMap[platform];
  const a = archMap[arch];
  if (!p || !a) throw new Error(`Unsupported platform: ${platform}/${arch}`);
  return { platform: p, arch: a };
}

function getBinaryName(platform) {
  return platform.platform === 'windows' ? 'zhizai.exe' : 'zhizai';
}

function getDownloadURL(platform) {
  const ext = platform.platform === 'windows' ? '.zip' : '.tar.gz';
  const asset = `zhizai-cli_${VERSION}_${platform.platform}_${platform.arch}${ext}`;
  return `https://github.com/${REPO}/releases/download/v${VERSION}/${asset}`;
}

function getWindowsExtractArgs(archivePath, destinationPath) {
  return [
    '-NoProfile',
    '-Command',
    '& { Expand-Archive -LiteralPath $args[0] -DestinationPath $args[1] -Force }',
    archivePath,
    destinationPath,
  ];
}

async function installArchive({ platform, binDir, binaryName, binaryPath, url, tmpFile }) {
  try {
    await download(url, tmpFile);
    await verifyChecksum(url, path.basename(url), tmpFile);
    if (platform.platform === 'windows') {
      run('powershell', getWindowsExtractArgs(tmpFile, binDir));
    } else {
      run('tar', ['-xzf', tmpFile, '-C', binDir, binaryName]);
    }

    if (!fs.existsSync(binaryPath)) {
      throw new Error(`Binary missing after extract: ${binaryPath}`);
    }
    // chmod is meaningful on Unix; Windows may not need it.
    if (platform.platform !== 'windows') {
      fs.chmodSync(binaryPath, 0o755);
    }
    console.log(`zhizai installed at ${binaryPath}`);
  } finally {
    try { fs.unlinkSync(tmpFile); } catch (_) {}
  }
}

async function main() {
  const platform = getPlatform();
  const binDir = path.join(__dirname, '..', 'bin');
  const binaryName = getBinaryName(platform);
  const binaryPath = path.join(binDir, binaryName);
  const url = getDownloadURL(platform);
  // Expand-Archive on Windows requires a .zip extension; tar.gz similarly.
  const archiveExt = platform.platform === 'windows' ? '.zip' : '.tar.gz';
  const tmpFile = path.join(os.tmpdir(), `zhizai-download-${Date.now()}${archiveExt}`);

  if (fs.existsSync(binaryPath)) {
    try {
      const result = spawnSync(binaryPath, ['version'], { encoding: 'utf8' });
      const out = (result.stdout || '').trim();
      if (result.status === 0 && out.includes(VERSION)) {
        console.log(`zhizai v${VERSION} already installed, skipping download.`);
        return;
      }
    } catch (_) {}
  }

  console.log(`Downloading zhizai v${VERSION} for ${platform.platform}/${platform.arch}...`);
  console.log(`URL: ${url}`);

  fs.mkdirSync(binDir, { recursive: true });
  await installArchive({ platform, binDir, binaryName, binaryPath, url, tmpFile });
}

function run(command, args) {
  const result = spawnSync(command, args, { stdio: 'inherit' });
  if (result.error) throw result.error;
  if (result.status !== 0) throw new Error(`${command} exited with status ${result.status}`);
}

function download(url, destination, redirects = 0) {
  if (redirects > 5) return Promise.reject(new Error('Too many redirects'));
  return new Promise((resolve, reject) => {
    const request = https.get(url, { headers: { 'User-Agent': '@zhizai/cli installer' } }, response => {
      if (response.statusCode >= 300 && response.statusCode < 400 && response.headers.location) {
        response.resume();
        return resolve(download(new URL(response.headers.location, url).toString(), destination, redirects + 1));
      }
      if (response.statusCode !== 200) {
        response.resume();
        return reject(new Error(`HTTP ${response.statusCode}: ${url}`));
      }
      const output = fs.createWriteStream(destination, { mode: 0o600 });
      response.pipe(output);
      output.on('finish', () => output.close(resolve));
      output.on('error', reject);
    });
    request.on('error', reject);
  });
}

async function verifyChecksum(assetURL, assetName, archivePath) {
  const checksumPath = `${archivePath}.checksums`;
  try {
    await download(new URL('checksums.txt', assetURL).toString(), checksumPath);
    const line = fs.readFileSync(checksumPath, 'utf8').split(/\r?\n/).find(value => {
      const fields = value.trim().split(/\s+/);
      return fields.length === 2 && fields[1].replace(/^\*/, '') === assetName;
    });
    if (!line) throw new Error(`Checksum for ${assetName} is missing`);
    const expected = line.trim().split(/\s+/)[0].toLowerCase();
    if (!/^[a-f0-9]{64}$/.test(expected)) throw new Error(`Checksum for ${assetName} is invalid`);
    const actual = crypto.createHash('sha256').update(fs.readFileSync(archivePath)).digest('hex');
    if (actual !== expected) throw new Error(`Checksum mismatch for ${assetName}`);
  } finally {
    try { fs.unlinkSync(checksumPath); } catch (_) {}
  }
}

if (require.main === module) {
  main().catch(err => {
    console.error('Failed to install zhizai:', err.message);
    process.exitCode = 1;
  });
}
