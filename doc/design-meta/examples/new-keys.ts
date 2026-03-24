import { KindOfTextNode, OperationStatus } from "./common";

type ChildParam = {
  localKeyId: string;
  kind: KindOfTextNode;
  flags?: string[];
  language?: string;
};

type NewKeyParams = {
  localKeyId: string;
  parentKeyId: string;
  secureParentKeyId: string;

  kind: KindOfTextNode;
  flags?: string[];
  language?: string;
  children: ChildParam[];
};

type SuggestedChildParam = {
  localKeyId: string;
  keyId: string;
  secureKeyId: string;

  kind: KindOfTextNode;
  flags?: string[];
  language?: string;
  status: OperationStatus;
};

type SuggestedNewKeyParams = {
  localKeyId: string;
  keyId: string;
  secureKeyId: string;
  kind: KindOfTextNode;
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
