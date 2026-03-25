import { KeyParams, KindOfTextNode, OperationStatus } from "./common";

type ChildParam = {
  localKeyId: string;
  kind: KindOfTextNode;
  flags?: string[];
  language?: string;
};

type NewKeyParams = {
  key: KeyParams;
  flags?: string[];
  language?: string;
  children: ChildParam[];
};

type SuggestedChildParam = {
  key: KeyParams;
  flags?: string[];
  language?: string;
  status: OperationStatus;
};

type SuggestedNewKeyParams = {
  key: KeyParams;
  flags?: string[];
  language?: string;
  status: OperationStatus;
  children: SuggestedChildParam[];
};

type NewKeysRequest = {
  newkeys: NewKeyParams[];
};

type NewKeysResponse = {
  id: string;
  newKeys: SuggestedNewKeyParams[];
};
