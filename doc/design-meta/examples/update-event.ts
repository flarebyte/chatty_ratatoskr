import type { KeyParams, OperationStatus } from './common';

type OnUpdateEvent = {
  id: string;
  eventList: [KeyParams, OperationStatus][];
};

export interface OnUpdateEventApi {
  onUpdate(): OnUpdateEvent;
}
