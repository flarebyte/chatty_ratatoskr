package httpapi

import (
	"fmt"
	"net/http"

	"github.com/flarebyte/chatty-ratatoskr/internal/yggkey"
)

type CreateAPI struct {
	generateID        func() string
	payloadLimitBytes int64
}

type childParam struct {
	LocalKeyID   string `json:"localKeyId"`
	ExpectedKind string `json:"expectedKind"`
}

type newKeyParams struct {
	Key          keyParams    `json:"key"`
	ExpectedKind string       `json:"expectedKind"`
	Children     []childParam `json:"children"`
}

type newKeysRequest struct {
	ID      string         `json:"id,omitempty"`
	RootKey keyParams      `json:"rootKey"`
	NewKeys []newKeyParams `json:"newKeys"`
}

type childStatusResult struct {
	Key     keyParams `json:"key"`
	Status  string    `json:"status"`
	Message string    `json:"message,omitempty"`
}

type suggestedNewKeyParams struct {
	Key      keyParams           `json:"key"`
	Status   string              `json:"status"`
	Message  string              `json:"message,omitempty"`
	Children []childStatusResult `json:"children"`
}

type newKeysResponseData struct {
	RootKey keyParams               `json:"rootKey"`
	NewKeys []suggestedNewKeyParams `json:"newKeys"`
}

func NewCreateAPI() *CreateAPI {
	return &CreateAPI{
		generateID:        func() string { return "generated" },
		payloadLimitBytes: defaultHTTPPayloadLimitBytes,
	}
}

func NewCreateAPIWithGenerator(generateID func() string) *CreateAPI {
	api := NewCreateAPI()
	if generateID != nil {
		api.generateID = generateID
	}
	return api
}

func NewCreateAPIWithOptions(generateID func() string, payloadLimitBytes int64) *CreateAPI {
	api := NewCreateAPI()
	if generateID != nil {
		api.generateID = generateID
	}
	if payloadLimitBytes > 0 {
		api.payloadLimitBytes = payloadLimitBytes
	}
	return api
}

func (api *CreateAPI) Register(mux *http.ServeMux) {
	mux.HandleFunc("/create", api.handleCreate)
}

func (api *CreateAPI) handleCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.NotFound(w, r)
		return
	}

	var req newKeysRequest
	if err := decodeJSONWithLimit(r, &req, api.payloadLimitBytes); err != nil {
		writeJSON(w, statusForDecodeError(err), invalidEnvelopeWithID(req.ID, api.generateID, messageForDecodeError(err)))
		return
	}
	if writeForcedStatusEnvelope(w, req.ID, api.generateID, req.RootKey.SecureKeyID) {
		return
	}
	rootParsed, ok := parseRootKeyOrWriteInvalid(w, req.ID, api.generateID, req.RootKey)
	if !ok {
		return
	}

	items := make([]suggestedNewKeyParams, 0, len(req.NewKeys))
	for _, item := range req.NewKeys {
		items = append(items, api.suggestNewKey(rootParsed, req.RootKey, item))
	}

	writeJSON(w, http.StatusOK, responseEnvelope[newKeysResponseData]{
		ID:     responseIDWithGenerator(req.ID, api.generateID),
		Status: "ok",
		Data: newKeysResponseData{
			RootKey: rootKeyResponse(req.RootKey, rootParsed),
			NewKeys: items,
		},
	})
}

func (api *CreateAPI) suggestNewKey(rootParsed yggkey.ParsedKey, root keyParams, item newKeyParams) suggestedNewKeyParams {
	result := suggestedNewKeyParams{
		Key: keyParams{
			KeyID:       item.Key.KeyID,
			SecureKeyID: item.Key.SecureKeyID,
			LocalKeyID:  item.Key.LocalKeyID,
		},
		Status:   "ok",
		Children: []childStatusResult{},
	}

	parentParsed, err := yggkey.Parse(item.Key.KeyID)
	if err != nil {
		result.Status = "invalid"
		result.Message = err.Error()
		return result
	}
	if !isDescendant(rootParsed.Canonical, item.Key.KeyID) && item.Key.KeyID != rootParsed.Canonical {
		result.Status = "invalid"
		result.Message = "invalid key: create parent must equal or descend from the requested root"
		return result
	}
	if !isCreatableContainerKind(item.ExpectedKind) {
		result.Status = "invalid"
		result.Message = fmt.Sprintf("invalid expectedKind: unsupported create kind %q", item.ExpectedKind)
		return result
	}

	newKeyID := item.Key.KeyID + ":" + item.ExpectedKind + ":" + api.generateID()
	newParsed, parseErr := yggkey.Parse(newKeyID)
	if parseErr != nil {
		result.Status = "invalid"
		result.Message = parseErr.Error()
		return result
	}

	result.Key = keyParams{
		KeyID:       newKeyID,
		SecureKeyID: chooseSecureKeyID(item.Key.SecureKeyID, root.SecureKeyID),
		LocalKeyID:  item.Key.LocalKeyID,
		Kind:        derivedKindParams(newParsed),
	}
	if item.Key.LocalKeyID == "" {
		result.Key.LocalKeyID = item.ExpectedKind
	}

	children := make([]childStatusResult, 0, len(item.Children))
	for _, child := range item.Children {
		childResult := childStatusResult{
			Key: keyParams{
				LocalKeyID:  child.LocalKeyID,
				SecureKeyID: chooseSecureKeyID(item.Key.SecureKeyID, root.SecureKeyID),
			},
			Status: "ok",
		}
		childKeyID, err := buildChildKeyID(newKeyID, child.ExpectedKind)
		if err != nil {
			childResult.Status = "invalid"
			childResult.Message = err.Error()
			children = append(children, childResult)
			continue
		}
		childParsed, parseErr := yggkey.Parse(childKeyID)
		if parseErr != nil {
			childResult.Status = "invalid"
			childResult.Message = parseErr.Error()
			children = append(children, childResult)
			continue
		}
		childResult.Key.KeyID = childKeyID
		childResult.Key.Kind = derivedKindParams(childParsed)
		children = append(children, childResult)
	}
	result.Children = children

	_ = parentParsed
	return result
}

func chooseSecureKeyID(primary, fallback string) string {
	if primary != "" {
		return primary
	}
	return fallback
}

func isCreatableContainerKind(kind string) bool {
	switch kind {
	case "note", "comment":
		return true
	default:
		return false
	}
}

func buildChildKeyID(parentKeyID, expectedKind string) (string, error) {
	switch expectedKind {
	case "text":
		return parentKeyID + ":text", nil
	case "thumbnail":
		return parentKeyID + ":thumbnail:_", nil
	case "language":
		return parentKeyID + ":language:_", nil
	default:
		return "", fmt.Errorf("invalid expectedKind: unsupported child kind %q", expectedKind)
	}
}
