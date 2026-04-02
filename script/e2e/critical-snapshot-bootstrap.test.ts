import { expect, test } from 'bun:test';

import { readCriticalFixture } from './helpers/fixtures';
import { jsonRequest } from './helpers/http';
import { startMockServer } from './helpers/mock-server';

test('critical snapshot bootstrap stays deterministic', async () => {
  const setSnapshotRequest = await readCriticalFixture(
    'set-snapshot.request.json',
  );
  const getSnapshotRequest = await readCriticalFixture(
    'get-snapshot.request.json',
  );
  const goldenFixture = await readCriticalFixture(
    'get-snapshot.response.golden.json',
  );
  const golden = `${JSON.stringify(JSON.parse(goldenFixture))}\n`;

  const runJourney = async (): Promise<string> => {
    const server = await startMockServer();
    try {
      await jsonRequest(
        'PUT',
        `${server.baseURL}/snapshot`,
        setSnapshotRequest,
      );
      return await jsonRequest(
        'GET',
        `${server.baseURL}/snapshot`,
        getSnapshotRequest,
      );
    } finally {
      await server.stop();
    }
  };

  const first = await runJourney();
  const second = await runJourney();

  expect(first).toBe(golden);
  expect(second).toBe(golden);
  expect(first).toBe(second);
}, 20000);
