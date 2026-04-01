import type { KeyParams, KeyValueParams } from './common';

/**
 * Archive is resource state, not an event operation.
 * Archived records are still emitted as `set` events with `--archived`
 * present in the payload options when appropriate.
 */
export type EventOperation = 'set' | 'snapshot-replaced';

export type EventEnvelope = {
  eventId: string;
  rootKey: KeyParams;
  operation: EventOperation;
  created: string;
  key?: KeyParams;
  keyValue?: KeyValueParams;
  snapshotVersion?: string;
};
