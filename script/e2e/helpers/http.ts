export type JSONResponse = {
  body: string;
  status: number;
};

const statusMarker = '\n__CHATTY_STATUS__:';

export async function jsonRequest(
  method: 'GET' | 'PUT' | 'POST',
  url: string,
  body: string,
): Promise<string> {
  const response = await jsonRequestWithStatus(method, url, body);
  return response.body;
}

export async function jsonRequestWithStatus(
  method: 'GET' | 'PUT' | 'POST',
  url: string,
  body: string,
): Promise<JSONResponse> {
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
      '--write-out',
      `${statusMarker}%{http_code}`,
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

  const markerIndex = stdout.lastIndexOf(statusMarker);
  if (markerIndex === -1) {
    throw new Error(
      `missing status marker in curl output for ${method} ${url}`,
    );
  }

  const responseBody = stdout.slice(0, markerIndex);
  const statusText = stdout.slice(markerIndex + statusMarker.length).trim();
  const status = Number(statusText);
  if (!Number.isInteger(status)) {
    throw new Error(`invalid status marker ${statusText} for ${method} ${url}`);
  }

  return {
    body: responseBody,
    status,
  };
}
