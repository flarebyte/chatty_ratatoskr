import { expect, test } from 'bun:test';

import { jsonRequestWithStatus } from './helpers/http';
import {
  isLoopbackUnavailable,
  type RunningServer,
  startMockServer,
} from './helpers/mock-server';

const rootKeyID = 'tenant:t8f3a1c2:group:g4b7d9e1:dashboard:d1e52f07';

test('create workflow preserves local keys and reports invalid kinds', async () => {
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
    const response = await jsonRequestWithStatus(
      'POST',
      `${server.baseURL}/create`,
      JSON.stringify({
        id: 'req-create-e2e-001',
        rootKey: {
          keyId: rootKeyID,
          secureKeyId: 'ok',
        },
        newKeys: [
          {
            key: {
              localKeyId: 'tmp-note-1',
              keyId: rootKeyID,
              secureKeyId: 'ok',
            },
            expectedKind: 'note',
            children: [
              { localKeyId: 'tmp-text-1', expectedKind: 'text' },
              { localKeyId: 'tmp-thumb-1', expectedKind: 'thumbnail' },
            ],
          },
          {
            key: {
              localKeyId: 'tmp-note-2',
              keyId: rootKeyID,
            },
            expectedKind: 'invalid-kind',
            children: [],
          },
        ],
      }),
    );

    expect(response.status).toBe(200);
    expect(JSON.parse(response.body)).toEqual({
      id: 'req-create-e2e-001',
      status: 'ok',
      data: {
        rootKey: {
          keyId: rootKeyID,
          secureKeyId: 'ok',
          kind: { hierarchy: ['dashboard'] },
        },
        newKeys: [
          {
            key: {
              localKeyId: 'tmp-note-1',
              keyId: `${rootKeyID}:note:generated`,
              secureKeyId: 'ok',
              kind: { hierarchy: ['dashboard', 'note'] },
            },
            status: 'ok',
            children: [
              {
                key: {
                  localKeyId: 'tmp-text-1',
                  keyId: `${rootKeyID}:note:generated:text`,
                  secureKeyId: 'ok',
                  kind: { hierarchy: ['dashboard', 'note', 'text'] },
                },
                status: 'ok',
              },
              {
                key: {
                  localKeyId: 'tmp-thumb-1',
                  keyId: `${rootKeyID}:note:generated:thumbnail:_`,
                  secureKeyId: 'ok',
                  kind: { hierarchy: ['dashboard', 'note', 'thumbnail'] },
                },
                status: 'ok',
              },
            ],
          },
          {
            key: {
              localKeyId: 'tmp-note-2',
              keyId: rootKeyID,
            },
            status: 'invalid',
            message:
              'invalid expectedKind: unsupported create kind "invalid-kind"',
            children: [],
          },
        ],
      },
    });
  } finally {
    await server.stop();
  }
}, 20000);
