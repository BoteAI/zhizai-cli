#!/usr/bin/env node
'use strict';

// Sync root package.json version into platform packages + optionalDependencies.

const fs = require('fs');
const path = require('path');
const { PLATFORMS } = require('./platform.js');

const ROOT = path.join(__dirname, '..');
const rootPkgPath = path.join(ROOT, 'package.json');
const rootPkg = JSON.parse(fs.readFileSync(rootPkgPath, 'utf8'));
const version = process.argv[2] || rootPkg.version;

if (!/^\d+\.\d+\.\d+([.-][\w.-]+)?$/.test(version)) {
  console.error(`invalid version: ${version}`);
  process.exit(1);
}

rootPkg.version = version;
rootPkg.optionalDependencies = Object.fromEntries(
  PLATFORMS.map(spec => [spec.pkg, version])
);
fs.writeFileSync(rootPkgPath, `${JSON.stringify(rootPkg, null, 2)}\n`);

for (const spec of PLATFORMS) {
  const pkgPath = path.join(ROOT, 'npm', 'platforms', spec.key, 'package.json');
  const pkg = JSON.parse(fs.readFileSync(pkgPath, 'utf8'));
  pkg.version = version;
  pkg.name = spec.pkg;
  fs.writeFileSync(pkgPath, `${JSON.stringify(pkg, null, 2)}\n`);
}

console.log(`synced npm package versions → ${version}`);
