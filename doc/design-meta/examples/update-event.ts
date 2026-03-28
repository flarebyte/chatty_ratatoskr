import type { KeyParams, OperationStatus } from './common';

type OnUpdateEvent = {
  id: string;
  eventList: [KeyParams, OperationStatus][];
};

export interface OnUpdateEventApi {
  send(key: KeyParams): void;
  onUpdate(): OnUpdateEvent;
}
