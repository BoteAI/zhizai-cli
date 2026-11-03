'use strict';

const fs = require('fs');
const path = require('path');
const os = require('os');

/** @typedef {{ key: string, pkg: string, goos: string, goarch: string, os: string[], cpu: string[], binary: string, archiveExt: string }} PlatformSpec */

/** @type {PlatformSpec[]} */
const PLATFORMS = [
  {
    key: 'darwin-arm64',
    pkg: '@zhizai/cli-darwin-arm64',
    goos: 'darwin',
    goarch: 'arm64',
    os: ['darwin'],
    cpu: ['arm64'],
    binary: 'zhizai',
    archiveExt: '.tar.gz',
  },
  {
    key: 'darwin-x64',
    pkg: '@zhizai/cli-darwin-x64',
    goos: 'darwin',
    goarch: 'amd64',
    os: ['darwin'],
    cpu: ['x64'],
    binary: 'zhizai',
    archiveExt: '.tar.gz',
  },
  {
    key: 'linux-arm64',
    pkg: '@zhizai/cli-linux-arm64',
    goos: 'linux',
    goarch: 'arm64',
    os: ['linux'],
    cpu: ['arm64'],
    binary: 'zhizai',
    archiveExt: '.tar.gz',
  },
  {
    key: 'linux-x64',
    pkg: '@zhizai/cli-linux-x64',
    goos: 'linux',
    goarch: 'amd64',
    os: ['linux'],
    cpu: ['x64'],
    binary: 'zhizai',
    archiveExt: '.tar.gz',
  },
  {
    key: 'windows-x64',
    pkg: '@zhizai/cli-windows-x64',
    goos: 'windows',
    goarch: 'amd64',
    os: ['win32'],
    cpu: ['x64'],
    binary: 'zhizai.exe',
    archiveExt: '.zip',
  },
  {
    key: 'windows-arm64',
    pkg: '@zhizai/cli-windows-arm64',
    goos: 'windows',
    goarch: 'arm64',
    os: ['win32'],
    cpu: ['arm64'],
    binary: 'zhizai.exe',
    archiveExt: '.zip',
  },
  // HarmonyOS / OpenHarmony Node reports process.platform === "openharmony".
  // Pure-Go static linux/arm64 ELF is used (CGO disabled); same approach as
  // several CLIs that treat OpenHarmony userspace as Linux-ABI compatible.
  {
    key: 'openharmony-arm64',
    pkg: '@zhizai/cli-openharmony-arm64',
    goos: 'openharmony',
    goarch: 'arm64',
    buildGoos: 'linux',
    buildGoarch: 'arm64',
    os: ['openharmony'],
    cpu: ['arm64'],
    binary: 'zhizai',
    archiveExt: '.tar.gz',
  },
];

function nodePlatformArch() {
  const platform = os.platform();
  const arch = os.arch();
  const platformMap = {
    darwin: 'darwin',
    linux: 'linux',
    win32: 'windows',
    openharmony: 'openharmony',
  };
  const archMap = { x64: 'x64', arm64: 'arm64' };
  const p = platformMap[platform];
  const a = archMap[arch];
  if (!p || !a) {
    throw new Error(`Unsupported platform: ${platform}/${arch}`);
  }
  return { platform: p, arch: a, key: `${p === 'windows' ? 'windows' : p}-${a}` };
}

function getCurrentPlatform() {
  const { key } = nodePlatformArch();
  const spec = PLATFORMS.find(item => item.key === key);
  if (!spec) {
    throw new Error(`Unsupported platform key: ${key}`);
  }
  return spec;
}

function getBinaryName(spec = getCurrentPlatform()) {
  return spec.binary;
}

function packageBinaryPath(pkgRoot, spec = getCurrentPlatform()) {
  return path.join(pkgRoot, 'bin', spec.binary);
}

function tryResolveOptionalPackageRoot(spec = getCurrentPlatform()) {
  try {
    const pkgJson = require.resolve(`${spec.pkg}/package.json`, {
      paths: [path.join(__dirname, '..'), process.cwd()],
    });
    return path.dirname(pkgJson);
  } catch (_) {}

  // Global/npm layout: sibling under the same @zhizai scope
  // .../node_modules/@zhizai/cli/scripts -> .../node_modules/@zhizai/cli-darwin-arm64
  const shortName = spec.pkg.split('/')[1];
  const siblingPkgJson = path.join(__dirname, '..', '..', shortName, 'package.json');
  if (fs.existsSync(siblingPkgJson)) {
    return path.dirname(siblingPkgJson);
  }

  return null;
}

function tryResolveOptionalBinary(spec = getCurrentPlatform()) {
  const root = tryResolveOptionalPackageRoot(spec);
  if (!root) return null;
  const binaryPath = packageBinaryPath(root, spec);
  if (!fs.existsSync(binaryPath)) return null;
  return binaryPath;
}

function localBinaryPath(binDir = path.join(__dirname, '..', 'bin'), spec = getCurrentPlatform()) {
  return path.join(binDir, spec.binary);
}

function releaseAssetName(version, spec) {
  return `zhizai-cli_${version}_${spec.goos}_${spec.goarch}${spec.archiveExt}`;
}

function githubDownloadURL(version, spec, repo = 'BoteAI/zhizai-cli') {
  return `https://github.com/${repo}/releases/download/v${version}/${releaseAssetName(version, spec)}`;
}

module.exports = {
  PLATFORMS,
  nodePlatformArch,
  getCurrentPlatform,
  getBinaryName,
  packageBinaryPath,
  tryResolveOptionalPackageRoot,
  tryResolveOptionalBinary,
  localBinaryPath,
  releaseAssetName,
  githubDownloadURL,
};
