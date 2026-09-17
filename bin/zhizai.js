#!/usr/bin/env node
// Platform-agnostic launcher for zhizai binary

'use strict';

const { spawn } = require('child_process');
const fs = require('fs');
const path = require('path');
const os = require('os');

const platform = os.platform();
const binaryName = platform === 'win32' ? 'zhizai.exe' : 'zhizai';
const binaryPath = path.join(__dirname, binaryName);

function failMissing() {
  console.error('无法启动智在记录：CLI 可执行文件缺失。');
  console.error('可能原因：安装中断或文件被移除。');
  console.error('请重新执行: npm install -g @zhizai/cli@latest');
  console.error('错误码：CLI_BINARY_MISSING');
  process.exit(1);
}

if (!fs.existsSync(binaryPath)) {
  failMissing();
}

const child = spawn(binaryPath, process.argv.slice(2), {
  stdio: 'inherit',
  windowsHide: true,
});

child.on('error', (err) => {
  if (err && err.code === 'ENOENT') {
    failMissing();
  }
  console.error('无法启动智在记录：', err.message);
  process.exit(1);
});

child.on('exit', (code, signal) => {
  if (signal) {
    process.kill(process.pid, signal);
  } else {
    process.exit(code);
  }
});
