import type { KeyParams, NodeKindExample, OperationStatus } from './common';
import type { KeyStatusResult, ResponseEnvelope } from './envelope';

type ChildParam = {
  localKeyId: string;
  expectedKind: NodeKindExample;
};

type NewKeyParams = {
  key: KeyParams;
  expectedKind: NodeKindExample;
  children: ChildParam[];
};

type SuggestedNewKeyParams = {
  key: KeyParams;
  status: OperationStatus;
  children: KeyStatusResult[];
};

type NewKeysRequest = {
  rootKey: KeyParams;
  newKeys: NewKeyParams[];
};

type NewKeysResponse = ResponseEnvelope<{
  rootKey: KeyParams;
  newKeys: SuggestedNewKeyParams[];
}>;

export interface NewKeysApi {
  createNewKeys(request: NewKeysRequest): NewKeysResponse;
}
