import type { OperationStatus } from './common';
import type { EventEnvelope } from './event-envelope';

export type SubscribeMessage = {
  kind: 'subscribe';
  // A client may send subscribe more than once to add further root keys
  // without reopening the WebSocket connection. Duplicate root keys are
  // normalized without error and the most recent subscription entry wins.
  rootKeys: string[];
};

export type UnsubscribeMessage = {
  kind: 'unsubscribe';
  rootKeys: string[];
};

export type PingMessage = {
  kind: 'ping';
};

export type ClientMessage =
  | SubscribeMessage
  | UnsubscribeMessage
  | PingMessage;

export type SubscribedMessage = {
  kind: 'subscribed';
  rootKeys: string[];
};

export type UnsubscribedMessage = {
  kind: 'unsubscribed';
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
};

export type ServerMessage =
  | SubscribedMessage
  | UnsubscribedMessage
  | EventMessage
  | StatusMessage
  | PongMessage;
