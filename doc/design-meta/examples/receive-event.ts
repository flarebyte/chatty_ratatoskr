import type { KeyParams, OperationStatus, UserParams } from "./common";
type Subscription = {
  id: string;
  user: UserParams;
  eventList: KeyParams[];
};

type EventResponse = {
  id: string;
  user: UserParams;
  eventList: [KeyParams, OperationStatus][];
};

export interface EventApi {
  registerUser(user: UserParams): [UserParams, OperationStatus];
  unregisterUser(user: UserParams): [UserParams, OperationStatus];
  subscribe(subscription: Subscription): EventResponse;
  unsubscribe(subscription: Subscription): EventResponse;
  send(user: UserParams, key: KeyParams): void;
  receiveUserUpdate(user: UserParams): EventResponse;
}
