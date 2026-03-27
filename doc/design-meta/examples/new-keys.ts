import type { KeyParams, NodeKind, OperationStatus } from './common';

type ChildParam = {
  localKeyId: string;
  expectedKind: NodeKind;
};

type NewKeyParams = {
  key: KeyParams;
  expectedKind: NodeKind;
  children: ChildParam[];
};

type SuggestedNewKeyParams = {
  key: KeyParams;
  status: OperationStatus;
  children: [KeyParams, OperationStatus][];
};

type NewKeysRequest = {
  rootKey: KeyParams;
  newkeys: NewKeyParams[];
};

type NewKeysResponse = {
  id: string;
  rootKey: KeyParams;
  newKeys: SuggestedNewKeyParams[];
};

export interface NewKeysApi {
  createNewKeys(request: NewKeysRequest): NewKeysResponse;
}
