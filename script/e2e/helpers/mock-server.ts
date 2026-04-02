import http from 'node:http';
import path from 'node:path';

import { repoRoot } from './paths';

export type RunningServer = {
  baseURL: string;
  stop: () => Promise<void>;
};

let nextPort = 19080;
let buildOnce: Promise<void> | undefined;

export async function startMockServer(): Promise<RunningServer> {
  await ensureBuiltBinary();

  const listen = `127.0.0.1:${nextPort}`;
  nextPort += 1;
  const binaryPath = path.join(repoRoot, '.e2e-bin', 'chatty');

  const proc = Bun.spawn([binaryPath, 'serve', '--listen', listen], {
    cwd: repoRoot,
    env: {
      ...process.env,
      PWD: repoRoot,
      GOCACHE: `${repoRoot}/.gocache`,
      GOMODCACHE: `${repoRoot}/.gomodcache`,
    },
    stdout: 'pipe',
    stderr: 'pipe',
  });

  await waitForReady(listen, proc);

  return {
    baseURL: `http://${listen}`,
    stop: async () => {
      proc.kill('SIGTERM');
      const exited = await Promise.race([
        proc.exited.then(() => true),
        sleep(2000).then(() => false),
      ]);
      if (!exited) {
        proc.kill('SIGKILL');
        await proc.exited;
      }
    },
  };
}

async function waitForReady(
  listen: string,
  proc: Bun.Subprocess,
): Promise<void> {
  const deadline = Date.now() + 10000;

  while (Date.now() < deadline) {
    const ready = await pingServer(listen);
    if (ready) {
      return;
    }
    await sleep(100);
  }

  const stdout = proc.stdout ? await new Response(proc.stdout).text() : '';
  const stderr = proc.stderr ? await new Response(proc.stderr).text() : '';
  throw new Error(
    `mock server did not become ready for ${listen}\nstdout=${stdout}\nstderr=${stderr}`,
  );
}

async function pingServer(listen: string): Promise<boolean> {
  return new Promise<boolean>((resolve) => {
    const req = http.request(
      {
        hostname: '127.0.0.1',
        port: Number(listen.split(':')[1]),
        path: '/snapshot',
        method: 'GET',
      },
      (res) => {
        res.resume();
        resolve(true);
      },
    );
    req.on('error', () => resolve(false));
    req.end();
  });
}

async function sleep(ms: number): Promise<void> {
  await new Promise((resolve) => setTimeout(resolve, ms));
}

async function ensureBuiltBinary(): Promise<void> {
  if (!buildOnce) {
    buildOnce = buildBinary();
  }
  await buildOnce;
}

async function buildBinary(): Promise<void> {
  const proc = Bun.spawn(['make', 'build-dev'], {
    cwd: repoRoot,
    env: {
      ...process.env,
      PWD: repoRoot,
    },
    stdout: 'inherit',
    stderr: 'inherit',
  });
  const exitCode = await proc.exited;
  if (exitCode !== 0) {
    throw new Error(`make build-dev failed with exit code ${exitCode}`);
  }
}
