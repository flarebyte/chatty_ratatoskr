import type { OperationStatus } from './common';
import type { EventEnvelope } from './event-envelope';

export type CommandMetadata = {
  // Optional client-provided correlation identifier echoed by the matching
  // acknowledgement or status response for the same command.
  id?: string;
};

export type SubscribeMessage = CommandMetadata & {
  kind: 'subscribe';
  // A client may send subscribe more than once to add further root keys
  // without reopening the WebSocket connection. Duplicate root keys are
  // normalized without error and the most recent subscription entry wins.
  rootKeys: string[];
};

export type UnsubscribeMessage = CommandMetadata & {
  kind: 'unsubscribe';
  rootKeys: string[];
};

export type PingMessage = CommandMetadata & {
  kind: 'ping';
};

export type ClientMessage = SubscribeMessage | UnsubscribeMessage | PingMessage;

export type SubscribedMessage = {
  kind: 'subscribed';
  id?: string;
  rootKeys: string[];
};

export type UnsubscribedMessage = {
  kind: 'unsubscribed';
  id?: string;
  rootKeys: string[];
};

export type EventMessage = {
  kind: 'event';
  event: EventEnvelope;
};

export type StatusMessage = {
  kind: 'status';
  id?: string;
  status: OperationStatus;
  message?: string;
};

export type PongMessage = {
  kind: 'pong';
  id?: string;
};

export type ServerMessage =
  | SubscribedMessage
  | UnsubscribedMessage
  | EventMessage
  | StatusMessage
  | PongMessage;
