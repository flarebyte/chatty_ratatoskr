import { Buffer } from 'node:buffer';
import net from 'node:net';

export type WebSocketSession = {
  close: () => Promise<void>;
  readJSON: (timeoutMs?: number) => Promise<unknown>;
  sendJSON: (value: unknown) => void;
};

type HandshakeState = {
  done: boolean;
  rejected: boolean;
};

type ReadFrameResult = {
  bytes: number;
  opcode: number;
  payload: Buffer;
};

const handshakeSuffix = '\r\n\r\n';
const maskKey = Buffer.from([0x63, 0x68, 0x61, 0x74]);

export async function startWebSocketSession(
  url: string,
): Promise<WebSocketSession> {
  const parsed = new URL(url);
  const host = parsed.hostname;
  const port = Number(parsed.port);
  const path = parsed.pathname + parsed.search;
  const socket = net.createConnection({ host, port });
  const queue: string[] = [];
  const waiters: Array<(value: string) => void> = [];
  const handshake: HandshakeState = { done: false, rejected: false };
  let buffered = Buffer.alloc(0);
  let closePromiseResolve: (() => void) | undefined;

  const closePromise = new Promise<void>((resolve) => {
    closePromiseResolve = resolve;
  });

  socket.on('close', () => {
    closePromiseResolve?.();
  });

  socket.on('data', (chunk) => {
    buffered = Buffer.concat([buffered, chunk]);
    if (!handshake.done) {
      const index = buffered.indexOf(handshakeSuffix);
      if (index === -1) {
        return;
      }
      const headers = buffered.subarray(0, index).toString('utf8');
      buffered = buffered.subarray(index + handshakeSuffix.length);
      if (!headers.startsWith('HTTP/1.1 101')) {
        handshake.rejected = true;
        socket.destroy(
          new Error(`websocket handshake failed for ${url}: ${headers}`),
        );
        return;
      }
      handshake.done = true;
    }

    while (handshake.done) {
      const frame = tryReadFrame(buffered);
      if (!frame) {
        return;
      }
      buffered = buffered.subarray(frame.bytes);
      switch (frame.opcode) {
        case 0x1: {
          const payload = frame.payload.toString('utf8');
          const next = waiters.shift();
          if (next) {
            next(payload);
            continue;
          }
          queue.push(payload);
          continue;
        }
        case 0x8:
          socket.end();
          return;
        case 0x9:
          socket.write(encodeFrame(0xa, frame.payload, false));
          continue;
        default:
          continue;
      }
    }
  });

  await new Promise<void>((resolve, reject) => {
    const onConnect = () => {
      socket.write(handshakeRequest(host, port, path));
    };
    const onError = (error: Error) => {
      cleanup();
      reject(error);
    };
    const onData = () => {
      if (!handshake.done || handshake.rejected) {
        return;
      }
      cleanup();
      resolve();
    };
    const cleanup = () => {
      socket.removeListener('connect', onConnect);
      socket.removeListener('error', onError);
      socket.removeListener('data', onData);
    };
    socket.on('connect', onConnect);
    socket.on('error', onError);
    socket.on('data', onData);
  });

  return {
    close: async () => {
      if (socket.destroyed) {
        return;
      }
      socket.write(encodeFrame(0x8, Buffer.alloc(0), true));
      socket.end();
      await closePromise;
    },
    readJSON: async (timeoutMs = 3000) => {
      const text =
        queue.shift() ??
        (await new Promise<string>((resolve, reject) => {
          const timer = setTimeout(() => {
            const index = waiters.indexOf(resolve);
            if (index >= 0) {
              waiters.splice(index, 1);
            }
            reject(
              new Error(`timed out waiting for websocket message from ${url}`),
            );
          }, timeoutMs);

          waiters.push((value) => {
            clearTimeout(timer);
            resolve(value);
          });
        }));
      return JSON.parse(text);
    },
    sendJSON: (value: unknown) => {
      socket.write(
        encodeFrame(0x1, Buffer.from(JSON.stringify(value), 'utf8'), true),
      );
    },
  };
}

function handshakeRequest(host: string, port: number, path: string): string {
  return [
    `GET ${path || '/'} HTTP/1.1`,
    `Host: ${host}:${port}`,
    'Upgrade: websocket',
    'Connection: Upgrade',
    'Sec-WebSocket-Version: 13',
    `Sec-WebSocket-Key: ${Buffer.from('chatty-e2e-key!!').toString('base64')}`,
    '',
    '',
  ].join('\r\n');
}

function tryReadFrame(buffer: Buffer): ReadFrameResult | null {
  if (buffer.length < 2) {
    return null;
  }

  const first = buffer[0];
  const second = buffer[1];
  const opcode = first & 0x0f;
  const masked = (second & 0x80) !== 0;
  let offset = 2;
  let length = second & 0x7f;

  if (length === 126) {
    if (buffer.length < offset + 2) {
      return null;
    }
    length = buffer.readUInt16BE(offset);
    offset += 2;
  } else if (length === 127) {
    if (buffer.length < offset + 8) {
      return null;
    }
    const big = buffer.readBigUInt64BE(offset);
    if (big > BigInt(Number.MAX_SAFE_INTEGER)) {
      throw new Error('websocket frame too large for test helper');
    }
    length = Number(big);
    offset += 8;
  }

  const maskBytes = masked ? 4 : 0;
  if (buffer.length < offset + maskBytes + length) {
    return null;
  }

  let payload = buffer.subarray(
    offset + maskBytes,
    offset + maskBytes + length,
  );
  if (masked) {
    const mask = buffer.subarray(offset, offset + maskBytes);
    payload = unmaskPayload(payload, mask);
  }

  return {
    bytes: offset + maskBytes + length,
    opcode,
    payload,
  };
}

function encodeFrame(opcode: number, payload: Buffer, masked: boolean): Buffer {
  const header = [0x80 | opcode];
  let lengthField = Buffer.alloc(0);
  if (payload.length < 126) {
    header.push((masked ? 0x80 : 0) | payload.length);
  } else if (payload.length <= 0xffff) {
    header.push((masked ? 0x80 : 0) | 126);
    lengthField = Buffer.alloc(2);
    lengthField.writeUInt16BE(payload.length);
  } else {
    header.push((masked ? 0x80 : 0) | 127);
    lengthField = Buffer.alloc(8);
    lengthField.writeBigUInt64BE(BigInt(payload.length));
  }

  if (!masked) {
    return Buffer.concat([Buffer.from(header), lengthField, payload]);
  }

  const maskedPayload = Buffer.alloc(payload.length);
  for (let index = 0; index < payload.length; index += 1) {
    maskedPayload[index] = payload[index] ^ maskKey[index % maskKey.length];
  }
  return Buffer.concat([
    Buffer.from(header),
    lengthField,
    maskKey,
    maskedPayload,
  ]);
}

function unmaskPayload(payload: Buffer, mask: Buffer): Buffer {
  const out = Buffer.alloc(payload.length);
  for (let index = 0; index < payload.length; index += 1) {
    out[index] = payload[index] ^ mask[index % mask.length];
  }
  return out;
}
