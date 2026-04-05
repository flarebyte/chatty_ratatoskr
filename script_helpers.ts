import { spawn } from 'node:child_process';
import { promises as fs } from 'node:fs';

export async function readVersionFromProjectYAML(): Promise<string> {
  const text = await fs.readFile('main.project.yaml', 'utf8');
  const match = text.match(/^\s*version:\s*([^\s#]+)\s*$/m);
  if (!match) {
    throw new Error('version not found in main.project.yaml');
  }
  return match[1];
}

export async function runChecked(
  cmd: string[],
  opts: { cwd?: string; env?: Record<string, string | undefined> } = {},
): Promise<void> {
  await new Promise<void>((resolve, reject) => {
    const proc = spawn(cmd[0], cmd.slice(1), {
      cwd: opts.cwd,
      env: opts.env,
      stdio: 'inherit',
    });

    proc.on('error', reject);
    proc.on('exit', (code) => {
      if (code === 0) {
        resolve();
        return;
      }
      reject(new Error(`command failed: ${cmd.join(' ')} (exit=${code})`));
    });
  });
}
