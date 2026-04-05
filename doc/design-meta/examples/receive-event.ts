import type { OperationStatus, PrincipalParams } from './common';
import type { EventEnvelope } from './event-envelope';
import type {
  ClientMessage,
  EventMessage,
  ServerMessage,
  StatusMessage,
  SubscribedMessage,
  SubscribeMessage,
  UnsubscribedMessage,
  UnsubscribeMessage,
} from './websocket-messages';

type Subscription = {
  id: string;
  principal: PrincipalParams;
  rootKeys: string[];
};

type EventResponse = {
  id: string;
  principal: PrincipalParams;
  eventList: [EventEnvelope, OperationStatus][];
};

export interface EventApi {
  registerPrincipal(
    principal: PrincipalParams,
  ): [PrincipalParams, OperationStatus];
  // Unregistering a principal clears all active subscriptions for that principal.
  unregisterPrincipal(
    principal: PrincipalParams,
  ): [PrincipalParams, OperationStatus];
  subscribe(subscription: Subscription): EventResponse;
  // Unsubscribing a key that is not currently subscribed is a no-op and does not raise an error.
  unsubscribe(subscription: Subscription): EventResponse;
  receivePrincipalUpdate(principal: PrincipalParams): EventResponse;
}

export interface WebSocketEventApi {
  onClientMessage(message: ClientMessage): ServerMessage | EventMessage;
  // Repeated subscribe messages extend the active root-key set for the connection.
  // Root subscriptions are predefined and apply to the full readable descendant subtree.
  // A subscribe request for a non-allowed root should return a status message with invalid.
  // When present, the client command id should be echoed by the matching reply.
  // Duplicate root keys are normalized and the most recent entry wins.
  subscribe(message: SubscribeMessage): SubscribedMessage | StatusMessage;
  // A valid unsubscribe request should return unsubscribed even when some keys were not active.
  unsubscribe(message: UnsubscribeMessage): UnsubscribedMessage | StatusMessage;
  // Closing the connection clears all active subscriptions tied to that connection.
  disconnect(principal: PrincipalParams): void;
}

export type EventHandlingRule = {
  operation: 'set' | 'snapshot-replaced';
  clientAction: string;
};

export const eventHandlingRules: EventHandlingRule[] = [
  {
    operation: 'set',
    clientAction:
      'Upsert the record locally. If options include --archived, treat archive as record state rather than a delete operation.',
  },
  {
    operation: 'snapshot-replaced',
    clientAction:
      'This is emitted after setSnapshot. Refetch the authoritative snapshot for the root key and replace the local baseline.',
  },
];

export interface EventProducerApi {
  emit(event: EventEnvelope): void;
  receivePrincipalUpdate(principal: PrincipalParams): EventResponse;
}
