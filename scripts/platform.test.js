'use strict';

const { test } = require('node:test');
const assert = require('node:assert/strict');
const {
  PLATFORMS,
  getCurrentPlatform,
  releaseAssetName,
  githubDownloadURL,
  tryResolveOptionalBinary,
} = require('./platform.js');

test('PLATFORMS cover expected npm package names', () => {
  const names = PLATFORMS.map(p => p.pkg).sort();
  assert.deepEqual(names, [
    '@zhizai/cli-darwin-arm64',
    '@zhizai/cli-darwin-x64',
    '@zhizai/cli-linux-arm64',
    '@zhizai/cli-linux-x64',
    '@zhizai/cli-openharmony-arm64',
    '@zhizai/cli-windows-arm64',
    '@zhizai/cli-windows-x64',
  ].sort());
});

test('getCurrentPlatform returns a known key for this host', () => {
  const spec = getCurrentPlatform();
  assert.ok(PLATFORMS.some(p => p.key === spec.key));
  assert.match(spec.pkg, /^@zhizai\/cli-/);
});

test('releaseAssetName uses goarch naming', () => {
  const darwinArm = PLATFORMS.find(p => p.key === 'darwin-arm64');
  const darwinX64 = PLATFORMS.find(p => p.key === 'darwin-x64');
  const win = PLATFORMS.find(p => p.key === 'windows-x64');
  const winArm = PLATFORMS.find(p => p.key === 'windows-arm64');
  assert.equal(releaseAssetName('0.0.10', darwinArm), 'zhizai-cli_0.0.10_darwin_arm64.tar.gz');
  assert.equal(releaseAssetName('0.0.10', darwinX64), 'zhizai-cli_0.0.10_darwin_amd64.tar.gz');
  assert.equal(releaseAssetName('0.0.10', win), 'zhizai-cli_0.0.10_windows_amd64.zip');
  assert.equal(releaseAssetName('0.0.10', winArm), 'zhizai-cli_0.0.10_windows_arm64.zip');
  const ohos = PLATFORMS.find(p => p.key === 'openharmony-arm64');
  assert.equal(releaseAssetName('0.0.10', ohos), 'zhizai-cli_0.0.10_openharmony_arm64.tar.gz');
  assert.equal(ohos.buildGoos, 'linux');
  assert.deepEqual(ohos.os, ['openharmony']);
});

test('githubDownloadURL points at GitHub releases', () => {
  const spec = PLATFORMS[0];
  assert.equal(
    githubDownloadURL('1.2.3', spec),
    `https://github.com/BoteAI/zhizai-cli/releases/download/v1.2.3/${releaseAssetName('1.2.3', spec)}`
  );
});

test('tryResolveOptionalBinary returns null when package missing', () => {
  const missing = {
    ...getCurrentPlatform(),
    pkg: '@zhizai/cli-does-not-exist-xyz',
  };
  assert.equal(tryResolveOptionalBinary(missing), null);
});
