import type { KeyParams, KeyValueParams, OperationStatus } from './common';

export type ResponseEnvelope<T> = {
  id: string;
  status: OperationStatus;
  message?: string;
  data: T;
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
