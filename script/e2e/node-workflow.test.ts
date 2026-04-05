import { expect, test } from 'bun:test';

import { jsonRequestWithStatus } from './helpers/http';
import {
  isLoopbackUnavailable,
  type RunningServer,
  startMockServer,
} from './helpers/mock-server';

const rootKeyID = 'tenant:t8f3a1c2:group:g4b7d9e1:dashboard:d1e52f07';

test('node write and read workflow works end to end', async () => {
  let server: RunningServer | undefined;
  try {
    server = await startMockServer();
  } catch (error) {
    if (isLoopbackUnavailable(error)) {
      return;
    }
    throw error;
  }

  try {
    const setResponse = await jsonRequestWithStatus(
      'PUT',
      `${server.baseURL}/node`,
      JSON.stringify({
        id: 'req-set-node-e2e-001',
        rootKey: {
          keyId: rootKeyID,
          secureKeyId: 'ok',
        },
        keyValueList: [
          {
            key: {
              keyId: rootKeyID,
              secureKeyId: 'ok',
              version: 'v1',
            },
            value: 'root-is-not-a-node-child',
          },
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

    expect(setResponse.status).toBe(200);
    const setBody = JSON.parse(setResponse.body);
    expect(setBody.status).toBe('ok');
    expect(setBody.data.keyList).toEqual([
      {
        key: {
          keyId: rootKeyID,
          secureKeyId: 'ok',
          version: 'v1',
          kind: { hierarchy: ['dashboard'] },
        },
        status: 'invalid',
        message:
          'invalid key: node entry must be a descendant of the requested root',
      },
      {
        key: {
          keyId: `${rootKeyID}:note:n7c401c2:text`,
          secureKeyId: 'ok',
          version: 'v1',
          kind: { hierarchy: ['dashboard', 'note', 'text'] },
        },
        status: 'ok',
      },
    ]);

    const getResponse = await jsonRequestWithStatus(
      'GET',
      `${server.baseURL}/node`,
      JSON.stringify({
        id: 'req-get-node-e2e-001',
        rootKey: {
          keyId: rootKeyID,
          secureKeyId: 'ok',
        },
        keyList: [
          {
            keyId: `${rootKeyID}:note:n7c401c2:text`,
            secureKeyId: 'ok',
          },
          {
            keyId: `${rootKeyID}:note:n7c401c2:missing:text`,
            secureKeyId: 'ok',
          },
        ],
      }),
    );

    expect(getResponse.status).toBe(200);
    const getBody = JSON.parse(getResponse.body);
    expect(getBody.status).toBe('ok');
    expect(getBody.data.keyValueList).toEqual([
      {
        keyValue: {
          key: {
            keyId: `${rootKeyID}:note:n7c401c2:text`,
            secureKeyId: 'ok',
            version: 'v1',
            kind: { hierarchy: ['dashboard', 'note', 'text'] },
          },
          value: 'hello world',
        },
        status: 'ok',
      },
      {
        keyValue: {
          key: {
            keyId: `${rootKeyID}:note:n7c401c2:missing:text`,
            secureKeyId: 'ok',
          },
        },
        status: 'invalid',
        message: 'invalid key: unsupported label "missing"',
      },
    ]);
  } finally {
    await server.stop();
  }
}, 20000);

test('node optimistic conflict rejection works end to end', async () => {
  let server: RunningServer | undefined;
  try {
    server = await startMockServer();
  } catch (error) {
    if (isLoopbackUnavailable(error)) {
      return;
    }
    throw error;
  }

  try {
    const firstWrite = await jsonRequestWithStatus(
      'PUT',
      `${server.baseURL}/node`,
      JSON.stringify({
        id: 'req-set-node-e2e-002',
        rootKey: {
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
            value: 'before',
          },
        ],
      }),
    );
    expect(firstWrite.status).toBe(200);

    const acceptedWrite = await jsonRequestWithStatus(
      'PUT',
      `${server.baseURL}/node`,
      JSON.stringify({
        id: 'req-set-node-e2e-003',
        rootKey: {
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
            value: 'after',
          },
        ],
      }),
    );
    expect(acceptedWrite.status).toBe(200);
    expect(acceptedWrite.body).toContain('"version":"v2"');

    const staleWrite = await jsonRequestWithStatus(
      'PUT',
      `${server.baseURL}/node`,
      JSON.stringify({
        id: 'req-set-node-e2e-004',
        rootKey: {
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
            value: 'stale',
          },
        ],
      }),
    );

    expect(staleWrite.status).toBe(409);
    expect(JSON.parse(staleWrite.body)).toEqual({
      id: 'req-set-node-e2e-004',
      status: 'outdated',
      data: {
        rootKey: {
          keyId: rootKeyID,
          secureKeyId: 'ok',
          kind: { hierarchy: ['dashboard'] },
        },
        keyList: [
          {
            key: {
              keyId: `${rootKeyID}:note:n7c401c2:text`,
              secureKeyId: 'ok',
              version: 'v1',
              kind: { hierarchy: ['dashboard', 'note', 'text'] },
            },
            status: 'outdated',
            message:
              'outdated version: write is not based on the latest stored version',
          },
        ],
      },
    });
  } finally {
    await server.stop();
  }
}, 20000);
