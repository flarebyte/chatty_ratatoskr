import type { OperationStatus, UserParams } from './common';
import type { EventEnvelope } from './event-envelope';

type Subscription = {
  id: string;
  user: UserParams;
  rootKeys: string[];
};

type EventResponse = {
  id: string;
  user: UserParams;
  eventList: [EventEnvelope, OperationStatus][];
};

export interface EventApi {
  registerUser(user: UserParams): [UserParams, OperationStatus];
  unregisterUser(user: UserParams): [UserParams, OperationStatus];
  subscribe(subscription: Subscription): EventResponse;
  unsubscribe(subscription: Subscription): EventResponse;
  receiveUserUpdate(user: UserParams): EventResponse;
}

export type EventHandlingRule = {
  operation: 'set' | 'snapshot-replaced';
  clientAction: string;
};

export const eventHandlingRules: EventHandlingRule[] = [
  {
    operation: 'set',
    clientAction: 'Upsert the record locally. If options include --archived, treat archive as record state.',
  },
  {
    operation: 'snapshot-replaced',
    clientAction: 'Refetch the authoritative snapshot for the root key and replace the local baseline.',
  },
];

export interface EventProducerApi {
  emit(event: EventEnvelope): void;
  receiveUserUpdate(user: UserParams): EventResponse;
}
