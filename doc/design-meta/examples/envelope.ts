import type { KeyParams, KeyValueParams, OperationStatus } from './common';

export type ResponseEnvelope<T> = {
  id: string;
  status: OperationStatus;
  message?: string;
  data: T;
};

export type RequestMetadata = {
  // Optional client-provided correlation identifier. If omitted, the server
  // should generate a response id.
  id?: string;
};

export type KeyStatusResult = {
  key: KeyParams;
  status: OperationStatus;
  message?: string;
};

export type KeyValueStatusResult = {
  keyValue: KeyValueParams;
  status: OperationStatus;
  message?: string;
};
