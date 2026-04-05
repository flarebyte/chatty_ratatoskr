import { expect, test } from 'bun:test';
import path from 'node:path';

import { jsonRequestWithStatus } from './helpers/http';
import {
  isLoopbackUnavailable,
  type RunningServer,
  startMockServer,
} from './helpers/mock-server';
import { repoRoot } from './helpers/paths';
import { startWebSocketSession } from './helpers/websocket';

const rootKeyID = 'tenant:t8f3a1c2:group:g4b7d9e1:dashboard:d1e52f07';

function eventsURL(baseURL: string): string {
  return `${baseURL.replace(/^http/, 'ws')}/events`;
}

function expectRecentTimestamp(value: unknown): void {
  expect(typeof value).toBe('string');
  expect(Number.isNaN(Date.parse(String(value)))).toBe(false);
}

test('websocket subscribe, ping, and unsubscribe work end to end', async () => {
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

  const ws = await startWebSocketSession(eventsURL(server.baseURL));
  try {
    ws.sendJSON({
      id: 'sub-e2e-001',
      kind: 'subscribe',
      rootKeys: [rootKeyID],
    });
    expect(await ws.readJSON()).toEqual({
      id: 'sub-e2e-001',
      kind: 'subscribed',
      rootKeys: [rootKeyID],
    });

    ws.sendJSON({
      id: 'ping-e2e-001',
      kind: 'ping',
    });
    expect(await ws.readJSON()).toEqual({
      id: 'ping-e2e-001',
      kind: 'pong',
    });

    ws.sendJSON({
      id: 'unsub-e2e-001',
      kind: 'unsubscribe',
      rootKeys: [rootKeyID, rootKeyID],
    });
    expect(await ws.readJSON()).toEqual({
      id: 'unsub-e2e-001',
      kind: 'unsubscribed',
    });
  } finally {
    await ws.close();
    await server.stop();
  }
}, 20000);

test('websocket event delivery works for node writes and snapshot replacement', async () => {
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

  const ws = await startWebSocketSession(eventsURL(server.baseURL));
  try {
    ws.sendJSON({
      id: 'sub-e2e-002',
      kind: 'subscribe',
      rootKeys: [rootKeyID],
    });
    expect(await ws.readJSON()).toEqual({
      id: 'sub-e2e-002',
      kind: 'subscribed',
      rootKeys: [rootKeyID],
    });

    const nodeResponse = await jsonRequestWithStatus(
      'PUT',
      `${server.baseURL}/node`,
      JSON.stringify({
        id: 'req-set-node-event-e2e-001',
        rootKey: {
          keyId: rootKeyID,
          secureKeyId: 'ok',
        },
        keyValueList: [
          {
            key: {
              keyId: `${rootKeyID}:note:n7c401c2:text`,
              secureKeyId: 'ok',
            },
            value: 'hello world',
          },
        ],
      }),
    );
    expect(nodeResponse.status).toBe(200);

    const setEvent = await ws.readJSON();
    expect(setEvent).toEqual({
      kind: 'event',
      event: {
        eventId: 'event-generated',
        rootKey: {
          keyId: rootKeyID,
          secureKeyId: 'ok',
          kind: { hierarchy: ['dashboard'] },
        },
        operation: 'set',
        created: expect.anything(),
        key: {
          keyId: `${rootKeyID}:note:n7c401c2:text`,
          secureKeyId: 'ok',
          version: 'v1',
          kind: { hierarchy: ['dashboard', 'note', 'text'] },
        },
        keyValue: {
          key: {
            keyId: `${rootKeyID}:note:n7c401c2:text`,
            secureKeyId: 'ok',
            version: 'v1',
            kind: { hierarchy: ['dashboard', 'note', 'text'] },
          },
          value: 'hello world',
        },
      },
    });
    expectRecentTimestamp(
      (setEvent as { event: { created: unknown } }).event.created,
    );

    const snapshotResponse = await jsonRequestWithStatus(
      'PUT',
      `${server.baseURL}/snapshot`,
      JSON.stringify({
        id: 'req-set-snapshot-event-e2e-001',
        key: {
          keyId: rootKeyID,
          secureKeyId: 'ok',
        },
        keyValueList: [
          {
            key: {
              keyId: `${rootKeyID}:note:n7c401c2:text`,
              secureKeyId: 'ok',
              version: 'v1',
            },
            value: 'hello world',
          },
        ],
      }),
    );
    expect(snapshotResponse.status).toBe(200);

    const snapshotEvent = await ws.readJSON();
    expect(snapshotEvent).toEqual({
      kind: 'event',
      event: {
        eventId: 'event-generated',
        rootKey: {
          keyId: rootKeyID,
          secureKeyId: 'ok',
          kind: { hierarchy: ['dashboard'] },
        },
        operation: 'snapshot-replaced',
        created: expect.anything(),
        snapshotVersion: 'snapshot-v1',
      },
    });
    expectRecentTimestamp(
      (snapshotEvent as { event: { created: unknown } }).event.created,
    );
  } finally {
    await ws.close();
    await server.stop();
  }
}, 20000);
