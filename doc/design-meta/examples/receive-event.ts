import type { OperationStatus, UserParams } from './common';
import type { EventEnvelope } from './event-envelope';
import type {
  ClientMessage,
  EventMessage,
  ServerMessage,
  StatusMessage,
  SubscribeMessage,
  SubscribedMessage,
  UnsubscribeMessage,
  UnsubscribedMessage,
} from './websocket-messages';

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
  // Unregistering a user clears all active subscriptions for that user.
  unregisterUser(user: UserParams): [UserParams, OperationStatus];
  subscribe(subscription: Subscription): EventResponse;
  // Unsubscribing a key that is not currently subscribed is a no-op and does not raise an error.
  unsubscribe(subscription: Subscription): EventResponse;
  receiveUserUpdate(user: UserParams): EventResponse;
}

export interface WebSocketEventApi {
  onClientMessage(message: ClientMessage): ServerMessage | EventMessage;
  // Repeated subscribe messages extend the active root-key set for the connection.
  // Root subscriptions are predefined and apply to the full readable descendant subtree.
  // A subscribe request for a non-allowed root should return a status message with invalid.
  // Duplicate root keys are normalized and the most recent entry wins.
  subscribe(message: SubscribeMessage): SubscribedMessage | StatusMessage;
  unsubscribe(message: UnsubscribeMessage): UnsubscribedMessage | StatusMessage;
  // Closing the connection clears all active subscriptions tied to that connection.
  disconnect(user: UserParams): void;
}

export type EventHandlingRule = {
  operation: 'set' | 'snapshot-replaced';
  clientAction: string;
};

export const eventHandlingRules: EventHandlingRule[] = [
  {
    operation: 'set',
    clientAction: 'Upsert the record locally. If options include --archived, treat archive as record state rather than a delete operation.',
  },
  {
    operation: 'snapshot-replaced',
    clientAction: 'This is emitted after setSnapshot. Refetch the authoritative snapshot for the root key and replace the local baseline.',
  },
];

export interface EventProducerApi {
  emit(event: EventEnvelope): void;
  receiveUserUpdate(user: UserParams): EventResponse;
}
