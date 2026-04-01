import type { KeyParams, KeyValueParams } from './common';

/**
 * Archive is resource state, not an event operation.
 * Archived records are still emitted as `set` events with `--archived`
 * present in the payload options when appropriate.
 *
 * `set` must carry `rootKey`, `key`, `keyValue`, and `created`.
 * `snapshot-replaced` must carry `rootKey`, `snapshotVersion`, and `created`.
 * `snapshot-replaced` is emitted after `setSnapshot`, not after every normal `set`.
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
