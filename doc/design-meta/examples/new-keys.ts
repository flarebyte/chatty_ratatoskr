type NewKeyType = "";

type KindOfTextNode =
  | "note"
  | "comment"
  | "label"
  | "like"
  | "avatar"
  | "emoticon"
  | "style"
  | "table"
  | "image/jpeg"
  | "altText"
  | "thumbnail"
  | "url";

type SuggestionStatus = "ok" | "invalid" | "unauthorised";

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
  status: SuggestionStatus;
};

type SuggestedNewKeyParams = {
  localKeyId: string;
  keyId: string;
  secureKeyId: string;
  kind: KindOfTextNode;
  flags?: string[];
  language?: string;
  status: SuggestionStatus;
  children: SuggestedChildParam[];
};

type NewKeysRequest = {
  newkeys: NewKeyParams[];
};

type NewkeysResponse = {
  newKeys: SuggestedNewKeyParams[];
};
