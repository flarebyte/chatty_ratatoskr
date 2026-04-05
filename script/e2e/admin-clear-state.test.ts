import { expect, test } from 'bun:test';
import path from 'node:path';

import { readCriticalFixture } from './helpers/fixtures';
import { jsonRequest } from './helpers/http';
import {
  isLoopbackUnavailable,
  type RunningServer,
  startMockServer,
} from './helpers/mock-server';
import { repoRoot } from './helpers/paths';

test('admin clear-state resets snapshot data', async () => {
  const setSnapshotRequest = await readCriticalFixture(
    'set-snapshot.request.json',
  );
  const getSnapshotRequest = await readCriticalFixture(
    'get-snapshot.request.json',
  );

  let server: RunningServer | undefined;
  try {
    server = await startMockServer({
      configPath: path.join(repoRoot, 'testdata', 'config', 'basic.cue'),
    });
  } catch (error) {
    if (isLoopbackUnavailable(error)) {
      return;
    }
    throw error;
  }
  try {
    await jsonRequest('PUT', `${server.baseURL}/snapshot`, setSnapshotRequest);

    const before = JSON.parse(
      await jsonRequest(
        'GET',
        `${server.baseURL}/snapshot`,
        getSnapshotRequest,
      ),
    );
    expect(before.data.keyValueList.length).toBe(2);

    await jsonRequest(
      'PUT',
      `${server.baseURL}/admin/commands`,
      JSON.stringify({
        id: 'req-clear-001',
        commands: [
          {
            id: 'clear-state',
            comment: 'Clear all mock-server in-memory stores',
            arguments: ['clear-state'],
          },
        ],
      }),
    );

    const after = JSON.parse(
      await jsonRequest(
        'GET',
        `${server.baseURL}/snapshot`,
        getSnapshotRequest,
      ),
    );
    expect(after.data.keyValueList).toEqual([]);
  } finally {
    await server.stop();
  }
}, 20000);
