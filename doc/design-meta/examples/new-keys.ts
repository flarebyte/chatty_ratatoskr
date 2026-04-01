import type { KeyParams, NodeKindExample, OperationStatus } from './common';

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
  children: [KeyParams, OperationStatus][];
};

type NewKeysRequest = {
  rootKey: KeyParams;
  newKeys: NewKeyParams[];
};

type NewKeysResponse = {
  id: string;
  rootKey: KeyParams;
  newKeys: SuggestedNewKeyParams[];
};

export interface NewKeysApi {
  createNewKeys(request: NewKeysRequest): NewKeysResponse;
}
