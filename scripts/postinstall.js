#!/usr/bin/env node
// postinstall.js — downloads the zhizai binary for the current platform.

'use strict';

const fs = require('fs');
const path = require('path');
const os = require('os');
const http = require('http');
const https = require('https');
const crypto = require('crypto');
const { spawnSync } = require('child_process');

const pkg = require('../package.json');
const VERSION = pkg.version;
const REPO = 'BoteAI/zhizai-cli';

const DEFAULT_CONNECT_TIMEOUT_MS = 15000;
const DEFAULT_IDLE_TIMEOUT_MS = 30000;
const DEFAULT_TOTAL_TIMEOUT_MS = 5 * 60 * 1000;
const DEFAULT_RETRIES = 3;
const RETRY_BACKOFF_MS = [500, 1000, 2000];
const NON_RETRYABLE_STATUS = new Set([401, 403, 404]);

function sleep(ms) {
  return new Promise(resolve => setTimeout(resolve, ms));
}

function getProtocolModule(urlString) {
  const protocol = new URL(urlString).protocol;
  if (protocol === 'http:') return http;
  if (protocol === 'https:') return https;
  throw new Error(`Unsupported protocol: ${protocol}`);
}

function versionMatches(out, expected) {
  const m = String(out).trim().match(/^zhizai version\s+v?(\S+)/i);
  if (!m) return false;
  return m[1] === String(expected).replace(/^v/, '');
}

function parseVersionFromOutput(out, expectedVersion) {
  return versionMatches(out, expectedVersion);
}

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

function unlinkQuiet(filePath) {
  try { fs.unlinkSync(filePath); } catch (_) {}
}

function reportProgress(downloaded, contentLength, onProgress) {
  if (contentLength != null) {
    console.error(`Downloaded ${downloaded}/${contentLength} bytes`);
  } else {
    console.error(`Downloaded ${downloaded} bytes`);
  }
  if (onProgress) onProgress({ downloaded, contentLength });
}

function isNonRetryableError(err) {
  const statusCode = err && err.statusCode;
  if (statusCode != null && NON_RETRYABLE_STATUS.has(statusCode)) return true;
  const message = String((err && err.message) || err);
  return /HTTP (401|403|404)/.test(message);
}

function downloadOnce(url, destination, options, redirects = 0) {
  if (redirects > 5) return Promise.reject(new Error('Too many redirects'));

  const {
    connectTimeoutMs,
    idleTimeoutMs,
    totalTimeoutMs,
    onProgress,
  } = options;

  return new Promise((resolve, reject) => {
    let request = null;
    let response = null;
    let output = null;
    let connected = false;
    let downloaded = 0;
    let contentLength = null;
    let lastProgressTime = 0;
    let idleTimer = null;
    let totalTimer = null;
    let connectTimer = null;
    let settled = false;

    function finish(err, value) {
      if (settled) return;
      settled = true;
      clearTimeout(idleTimer);
      clearTimeout(totalTimer);
      clearTimeout(connectTimer);
      if (err) reject(err);
      else resolve(value);
    }

    function cleanupPartial() {
      unlinkQuiet(destination);
    }

    function destroyStreams() {
      if (response) {
        try { response.destroy(); } catch (_) {}
      }
      if (request) {
        try { request.destroy(); } catch (_) {}
      }
      if (output) {
        try { output.destroy(); } catch (_) {}
      }
    }

    function failTimeout(stage) {
      destroyStreams();
      cleanupPartial();
      finish(new Error(`${stage} timeout`));
    }

    function resetIdleTimer() {
      clearTimeout(idleTimer);
      idleTimer = setTimeout(() => failTimeout('Idle'), idleTimeoutMs);
    }

    function maybeReportProgress(force = false) {
      const now = Date.now();
      if (!force && now - lastProgressTime < 500) return;
      lastProgressTime = now;
      reportProgress(downloaded, contentLength, onProgress);
    }

    totalTimer = setTimeout(() => failTimeout('Total'), totalTimeoutMs);

    const protocol = getProtocolModule(url);
    request = protocol.get(url, { headers: { 'User-Agent': '@zhizai/cli installer' } }, res => {
      connected = true;
      clearTimeout(connectTimer);
      response = res;

      if (res.statusCode >= 300 && res.statusCode < 400 && res.headers.location) {
        res.resume();
        destroyStreams();
        clearTimeout(idleTimer);
        clearTimeout(totalTimer);
        clearTimeout(connectTimer);
        settled = true;
        downloadOnce(new URL(res.headers.location, url).toString(), destination, options, redirects + 1)
          .then(resolve, reject);
        return;
      }

      if (res.statusCode !== 200) {
        res.resume();
        destroyStreams();
        cleanupPartial();
        const err = new Error(`HTTP ${res.statusCode}: ${url}`);
        err.statusCode = res.statusCode;
        finish(err);
        return;
      }

      const lengthHeader = res.headers['content-length'];
      if (lengthHeader != null && lengthHeader !== '') {
        const parsed = Number(lengthHeader);
        if (!Number.isNaN(parsed)) contentLength = parsed;
      }

      output = fs.createWriteStream(destination, { mode: 0o600 });
      resetIdleTimer();

      res.on('data', chunk => {
        downloaded += chunk.length;
        resetIdleTimer();
        maybeReportProgress();
      });

      res.pipe(output);

      output.on('finish', () => {
        output.close(() => {
          maybeReportProgress(true);
          finish(null, undefined);
        });
      });

      output.on('error', err => {
        destroyStreams();
        cleanupPartial();
        finish(err);
      });

      res.on('error', err => {
        destroyStreams();
        cleanupPartial();
        finish(err);
      });
    });

    connectTimer = setTimeout(() => {
      if (!connected) failTimeout('Connect');
    }, connectTimeoutMs);

    request.on('socket', socket => {
      socket.once('connect', () => {
        connected = true;
        clearTimeout(connectTimer);
        resetIdleTimer();
      });
    });

    request.on('error', err => {
      destroyStreams();
      cleanupPartial();
      finish(err);
    });
  });
}

async function download(url, destination, options = {}, redirects = 0) {
  const opts = {
    connectTimeoutMs: DEFAULT_CONNECT_TIMEOUT_MS,
    idleTimeoutMs: DEFAULT_IDLE_TIMEOUT_MS,
    totalTimeoutMs: DEFAULT_TOTAL_TIMEOUT_MS,
    retries: DEFAULT_RETRIES,
    onProgress: null,
    ...options,
  };

  if (redirects > 5) throw new Error('Too many redirects');

  const maxRetries = opts.retries;
  let lastError;

  for (let attempt = 0; attempt <= maxRetries; attempt += 1) {
    try {
      return await downloadOnce(url, destination, opts, redirects);
    } catch (err) {
      lastError = err;
      if (isNonRetryableError(err) || attempt >= maxRetries) throw err;
      const delay = RETRY_BACKOFF_MS[Math.min(attempt, RETRY_BACKOFF_MS.length - 1)];
      await sleep(delay);
    }
  }

  throw lastError;
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
    if (platform.platform !== 'windows') {
      fs.chmodSync(binaryPath, 0o755);
    }
    console.log(`zhizai installed at ${binaryPath}`);
  } finally {
    unlinkQuiet(tmpFile);
  }
}

async function main() {
  const platform = getPlatform();
  const binDir = path.join(__dirname, '..', 'bin');
  const binaryName = getBinaryName(platform);
  const binaryPath = path.join(binDir, binaryName);
  const url = getDownloadURL(platform);
  const archiveExt = platform.platform === 'windows' ? '.zip' : '.tar.gz';
  const tmpFile = path.join(os.tmpdir(), `zhizai-download-${Date.now()}${archiveExt}`);

  if (fs.existsSync(binaryPath)) {
    try {
      const result = spawnSync(binaryPath, ['version'], { encoding: 'utf8' });
      const out = (result.stdout || '').trim();
      if (result.status === 0 && versionMatches(out, VERSION)) {
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
    unlinkQuiet(checksumPath);
  }
}

module.exports = {
  download,
  downloadOnce,
  versionMatches,
  parseVersionFromOutput,
  getPlatform,
  getBinaryName,
  getDownloadURL,
  verifyChecksum,
  installArchive,
};

if (require.main === module) {
  main().catch(err => {
    console.error('Failed to install zhizai:', err.message);
    process.exitCode = 1;
  });
}
