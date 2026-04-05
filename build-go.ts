#!/usr/bin/env bun

// Builds the chatty Go CLI for multiple platforms and writes checksums.

import crypto from 'node:crypto';
import { promises as fs } from 'node:fs';
import { readVersionFromProjectYAML, runChecked } from './script_helpers';

async function ensureDir(p: string): Promise<void> {
  await fs.mkdir(p, { recursive: true });
}

async function sha256File(filePath: string): Promise<string> {
  const hash = crypto.createHash('sha256');
  const file = Bun.file(filePath);
  const stream = file.stream();
  const reader = stream.getReader();
  while (true) {
    const { value, done } = await reader.read();
    if (done) break;
    if (value) hash.update(value);
  }
  return hash.digest('hex');
}

async function main() {
  const version = (await readVersionFromProjectYAML()).trim();
  if (!version)
    throw new Error('version not found in main.project.yaml (tags.version)');

  const commitFromGit =
    (
      await Bun.$`git rev-parse --short=12 HEAD`
        .quiet()
        .text()
        .catch(() => '')
    ).trim() || 'unknown';
  const commit = process.env.COMMIT || commitFromGit || 'unknown';
  const currentDate =
    process.env.DATE ?? new Date().toISOString().replace(/\.\d{3}Z$/, 'Z');

  const ldflags = [
    `-X main.Version=${version}`,
    `-X main.Commit=${commit}`,
    `-X main.Date=${currentDate}`,
  ].join(' ');

  const platforms = [
    { label: 'Linux (amd64)', os: 'linux', arch: 'amd64' },
    { label: 'Linux (arm64)', os: 'linux', arch: 'arm64' },
    { label: 'macOS (Apple Silicon)', os: 'darwin', arch: 'arm64' },
  ] as const;

  await ensureDir('build');

  const builtFiles: string[] = [];

  for (const p of platforms) {
    console.log(p.label);
    const env: Record<string, string> = { ...process.env } as Record<
      string,
      string
    >;
    env.GOOS = p.os;
    env.GOARCH = p.arch;
    if (p.os === 'darwin') {
      const macArch = p.arch === 'amd64' ? 'x86_64' : 'arm64';
      env.CGO_ENABLED = '1';
      env.CC = 'clang';
      env.CGO_CFLAGS = `-arch ${macArch}`;
      env.CGO_LDFLAGS = `-arch ${macArch}`;
      env.MACOSX_DEPLOYMENT_TARGET = env.MACOSX_DEPLOYMENT_TARGET || '11.0';
    }

    const out = `build/chatty-${p.os}-${p.arch}`;
    await runChecked(
      ['go', 'build', '-o', out, '-ldflags', ldflags, './cmd/chatty'],
      { env },
    );
    builtFiles.push(out);
  }

  // checksums (sha256), format: "<hex>  <path>" like shasum
  const lines: string[] = [];
  for (const f of builtFiles) {
    const digest = await sha256File(f);
    lines.push(`${digest}  ${f}`);
  }
  await fs.writeFile('build/checksums.txt', `${lines.join('\n')}\n`, 'utf8');
}

main().catch((err) => {
  console.error(err);
  process.exitCode = 1;
});
