'use strict';

const { test } = require('node:test');
const assert = require('node:assert/strict');
const http = require('http');
const fs = require('fs');
const os = require('os');
const path = require('path');
const { download, versionMatches } = require('./postinstall.js');

test('download idle timeout rejects when server sends no headers', async () => {
  const server = http.createServer(() => {
    // accept TCP but never write response headers
  });
  await new Promise(r => server.listen(0, '127.0.0.1', r));
  const { port } = server.address();
  const dest = path.join(os.tmpdir(), `zhizai-test-${Date.now()}.bin`);
  const start = Date.now();
  await assert.rejects(
    () => download(`http://127.0.0.1:${port}/hang`, dest, {
      connectTimeoutMs: 1000,
      idleTimeoutMs: 300,
      totalTimeoutMs: 60000,
      retries: 0,
    }),
    /Idle timeout/i
  );
  assert.ok(Date.now() - start < 5000, 'should fail on idle timeout, not total');
  server.close();
  try { fs.unlinkSync(dest); } catch (_) {}
});

test('download idle timeout rejects stalled body', async () => {
  const server = http.createServer((req, res) => {
    res.writeHead(200, { 'Content-Length': '100' });
    // never write body
  });
  await new Promise(r => server.listen(0, '127.0.0.1', r));
  const { port } = server.address();
  const dest = path.join(os.tmpdir(), `zhizai-test-${Date.now()}.bin`);
  await assert.rejects(
    () => download(`http://127.0.0.1:${port}/slow`, dest, {
      connectTimeoutMs: 1000,
      idleTimeoutMs: 200,
      totalTimeoutMs: 2000,
      retries: 0,
    }),
    /timeout|idle/i
  );
  server.close();
  try { fs.unlinkSync(dest); } catch (_) {}
});

test('download completes HTTP 200 response', async () => {
  const body = Buffer.from('hello-zhizai');
  const server = http.createServer((req, res) => {
    res.writeHead(200, { 'Content-Length': String(body.length) });
    res.end(body);
  });
  await new Promise(r => server.listen(0, '127.0.0.1', r));
  const { port } = server.address();
  const dest = path.join(os.tmpdir(), `zhizai-test-${Date.now()}.bin`);
  try {
    await download(`http://127.0.0.1:${port}/ok`, dest, {
      connectTimeoutMs: 1000,
      idleTimeoutMs: 2000,
      totalTimeoutMs: 5000,
      retries: 0,
    });
    assert.equal(fs.readFileSync(dest, 'utf8'), 'hello-zhizai');
  } finally {
    server.close();
    try { fs.unlinkSync(dest); } catch (_) {}
  }
});

test('download does not retry HTTP 404', async () => {
  let hits = 0;
  const server = http.createServer((req, res) => {
    hits += 1;
    res.writeHead(404);
    res.end('not found');
  });
  await new Promise(r => server.listen(0, '127.0.0.1', r));
  const { port } = server.address();
  const dest = path.join(os.tmpdir(), `zhizai-test-${Date.now()}.bin`);
  try {
    await assert.rejects(
      () => download(`http://127.0.0.1:${port}/missing`, dest, {
        connectTimeoutMs: 1000,
        idleTimeoutMs: 2000,
        totalTimeoutMs: 5000,
        retries: 3,
      }),
      /HTTP 404/
    );
    assert.equal(hits, 1);
  } finally {
    server.close();
    try { fs.unlinkSync(dest); } catch (_) {}
  }
});

test('versionMatches requires exact version', () => {
  assert.equal(versionMatches('zhizai version 1.2.3', '1.2.3'), true);
  assert.equal(versionMatches('zhizai version v1.2.3', '1.2.3'), true);
  assert.equal(versionMatches('zhizai version 1.2.30', '1.2.3'), false);
  assert.equal(versionMatches('prefix zhizai version 1.2.3', '1.2.3'), false);
});
