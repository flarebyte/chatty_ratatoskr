export async function jsonRequest(
  method: 'GET' | 'PUT',
  url: string,
  body: string,
): Promise<string> {
  const proc = Bun.spawn(
    [
      'curl',
      '-sS',
      '-X',
      method,
      url,
      '-H',
      'content-type: application/json',
      '--data-binary',
      body,
    ],
    {
      stdout: 'pipe',
      stderr: 'pipe',
    },
  );

  const exitCode = await proc.exited;
  const stdout = proc.stdout ? await new Response(proc.stdout).text() : '';
  const stderr = proc.stderr ? await new Response(proc.stderr).text() : '';

  if (exitCode !== 0) {
    throw new Error(`${method} ${url} failed with exit=${exitCode}: ${stderr}`);
  }
  return stdout;
}
