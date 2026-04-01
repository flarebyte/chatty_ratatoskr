import type { KeyParams, NodeKindExample, OperationStatus } from './common';
import type { KeyStatusResult, RequestMetadata, ResponseEnvelope } from './envelope';

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

type NewKeysRequest = RequestMetadata & {
  rootKey: KeyParams;
  newKeys: NewKeyParams[]; // processed independently and returned in request order
};

type NewKeysResponse = ResponseEnvelope<{
  rootKey: KeyParams;
  newKeys: SuggestedNewKeyParams[]; // every requested item should receive a corresponding per-item status
}>;

export interface NewKeysApi {
  createNewKeys(request: NewKeysRequest): NewKeysResponse;
}
