package httpapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/flarebyte/chatty-ratatoskr/internal/snapshot"
	"github.com/flarebyte/chatty-ratatoskr/internal/yggkey"
)

const defaultGeneratedResponseID = "generated"
const defaultHTTPPayloadLimitBytes int64 = 1 << 20

var errPayloadTooLarge = errors.New("payload too large")

type forcedStatus struct {
	httpStatus int
	status     string
	message    string
}

type SnapshotAPI struct {
	store             snapshot.Store
	generateID        func() string
	events            *EventsAPI
	payloadLimitBytes int64
}

type kindParams struct {
	Hierarchy []string `json:"hierarchy,omitempty"`
	Language  string   `json:"language,omitempty"`
}

type keyParams struct {
	LocalKeyID  string      `json:"localKeyId,omitempty"`
	KeyID       string      `json:"keyId"`
	SecureKeyID string      `json:"secureKeyId,omitempty"`
	Version     string      `json:"version,omitempty"`
	Kind        *kindParams `json:"kind,omitempty"`
}

type keyValueParams struct {
	Key   keyParams `json:"key"`
	Value string    `json:"value,omitempty"`
}

type setSnapshotRequest struct {
	ID           string           `json:"id,omitempty"`
	Key          keyParams        `json:"key"`
	KeyValueList []keyValueParams `json:"keyValueList"`
}

type getSnapshotRequest struct {
	ID  string    `json:"id,omitempty"`
	Key keyParams `json:"key"`
}

type responseEnvelope[T any] struct {
	ID      string `json:"id"`
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
	Data    T      `json:"data"`
}

type setSnapshotResponseData struct {
	Key keyParams `json:"key"`
}

type getSnapshotResponseData struct {
	Key          keyParams        `json:"key"`
	KeyValueList []keyValueParams `json:"keyValueList"`
}

func NewSnapshotAPI(store snapshot.Store) *SnapshotAPI {
	return &SnapshotAPI{
		store:             store,
		generateID:        func() string { return defaultGeneratedResponseID },
		payloadLimitBytes: defaultHTTPPayloadLimitBytes,
	}
}

func NewSnapshotAPIWithEvents(store snapshot.Store, events *EventsAPI) *SnapshotAPI {
	api := NewSnapshotAPI(store)
	api.events = events
	return api
}

func NewSnapshotAPIWithOptions(store snapshot.Store, events *EventsAPI, payloadLimitBytes int64) *SnapshotAPI {
	api := NewSnapshotAPI(store)
	api.events = events
	if payloadLimitBytes > 0 {
		api.payloadLimitBytes = payloadLimitBytes
	}
	return api
}

func NewSnapshotAPIWithGenerator(store snapshot.Store, generateID func() string) *SnapshotAPI {
	api := NewSnapshotAPI(store)
	if generateID != nil {
		api.generateID = generateID
	}
	return api
}

func NewSnapshotAPIWithLimit(store snapshot.Store, payloadLimitBytes int64) *SnapshotAPI {
	api := NewSnapshotAPI(store)
	if payloadLimitBytes > 0 {
		api.payloadLimitBytes = payloadLimitBytes
	}
	return api
}

func (api *SnapshotAPI) Register(mux *http.ServeMux) {
	mux.HandleFunc("/snapshot", api.handleSnapshot)
}

func (api *SnapshotAPI) handleSnapshot(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		api.handleGetSnapshot(w, r)
	case http.MethodPut:
		api.handleSetSnapshot(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (api *SnapshotAPI) handleSetSnapshot(w http.ResponseWriter, r *http.Request) {
	var req setSnapshotRequest
	if err := decodeJSONWithLimit(r, &req, api.payloadLimitBytes); err != nil {
		writeJSON(w, statusForDecodeError(err), api.invalidEnvelope(req.ID, messageForDecodeError(err)))
		return
	}
	rootParsed, root, ok := api.parseSnapshotRootOrWriteInvalid(w, req.ID, req.Key)
	if !ok {
		return
	}

	entries := make([]snapshot.KeyValue, 0, len(req.KeyValueList))
	for _, item := range req.KeyValueList {
		if _, err := yggkey.Parse(item.Key.KeyID); err != nil {
			writeJSON(w, http.StatusBadRequest, api.invalidEnvelope(req.ID, err.Error()))
			return
		}
		if !isDescendant(rootParsed.Canonical, item.Key.KeyID) {
			writeJSON(w, http.StatusBadRequest, api.invalidEnvelope(req.ID, "invalid key: snapshot entry must be a descendant of the requested root"))
			return
		}
		entries = append(entries, snapshot.KeyValue{
			Key: snapshot.Key{
				KeyID:       item.Key.KeyID,
				SecureKeyID: item.Key.SecureKeyID,
				Version:     item.Key.Version,
			},
			Value: item.Value,
		})
	}

	api.store.Replace(root, entries)
	if api.events != nil {
		api.events.EmitSnapshotReplaced(rootKeyResponse(req.Key, rootParsed), "snapshot-v1")
	}

	writeJSON(w, http.StatusOK, responseEnvelope[setSnapshotResponseData]{
		ID:     api.responseID(req.ID),
		Status: "ok",
		Data: setSnapshotResponseData{
			Key: rootKeyResponse(req.Key, rootParsed),
		},
	})
}

func (api *SnapshotAPI) handleGetSnapshot(w http.ResponseWriter, r *http.Request) {
	var req getSnapshotRequest
	if err := decodeJSONWithLimit(r, &req, api.payloadLimitBytes); err != nil {
		writeJSON(w, statusForDecodeError(err), api.invalidEnvelope(req.ID, messageForDecodeError(err)))
		return
	}
	rootParsed, root, ok := api.parseSnapshotRootOrWriteInvalid(w, req.ID, req.Key)
	if !ok {
		return
	}
	got := api.store.Get(root)

	keyValueList := make([]keyValueParams, 0, len(got.KeyValueList))
	for _, item := range got.KeyValueList {
		keyValueList = append(keyValueList, keyValueParams{
			Key: keyParams{
				KeyID:       item.Key.KeyID,
				SecureKeyID: item.Key.SecureKeyID,
				Version:     item.Key.Version,
			},
			Value: item.Value,
		})
	}

	writeJSON(w, http.StatusOK, responseEnvelope[getSnapshotResponseData]{
		ID:     api.responseID(req.ID),
		Status: "ok",
		Data: getSnapshotResponseData{
			Key:          rootKeyResponse(req.Key, rootParsed),
			KeyValueList: keyValueList,
		},
	})
}

func derivedKindParams(parsed yggkey.ParsedKey) *kindParams {
	kind := parsed.DerivedKind()
	return &kindParams{Hierarchy: kind.Hierarchy}
}

func decodeJSONWithLimit(r *http.Request, out any, limitBytes int64) error {
	var reader io.Reader = r.Body
	if limitBytes > 0 {
		data, err := io.ReadAll(io.LimitReader(r.Body, limitBytes+1))
		if err != nil {
			return err
		}
		if int64(len(data)) > limitBytes {
			return errPayloadTooLarge
		}
		reader = bytes.NewReader(data)
	}
	decoder := json.NewDecoder(reader)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(out); err != nil {
		if errors.Is(err, io.EOF) {
			return err
		}
		return err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return io.ErrUnexpectedEOF
	}
	return nil
}

func (api *SnapshotAPI) responseID(requestID string) string {
	return responseIDWithGenerator(requestID, api.generateID)
}

func (api *SnapshotAPI) invalidEnvelope(requestID, message string) responseEnvelope[map[string]any] {
	return responseEnvelope[map[string]any]{
		ID:      api.responseID(requestID),
		Status:  "invalid",
		Message: message,
		Data:    map[string]any{},
	}
}

func invalidEnvelopeWithID(requestID string, generateID func() string, message string) responseEnvelope[map[string]any] {
	return responseEnvelope[map[string]any]{
		ID:      responseIDWithGenerator(requestID, generateID),
		Status:  "invalid",
		Message: message,
		Data:    map[string]any{},
	}
}

func (api *SnapshotAPI) parseSnapshotRootOrWriteInvalid(w http.ResponseWriter, requestID string, key keyParams) (yggkey.ParsedKey, snapshot.Key, bool) {
	if writeForcedStatusEnvelope(w, requestID, api.generateID, key.SecureKeyID) {
		return yggkey.ParsedKey{}, snapshot.Key{}, false
	}
	parsed, ok := parseRootKeyOrWriteInvalid(w, requestID, api.generateID, key)
	if !ok {
		return yggkey.ParsedKey{}, snapshot.Key{}, false
	}
	return parsed, snapshot.Key{
		KeyID:       key.KeyID,
		SecureKeyID: key.SecureKeyID,
		Version:     key.Version,
	}, true
}

func rootKeyResponse(root keyParams, parsed yggkey.ParsedKey) keyParams {
	return keyParams{
		KeyID:       root.KeyID,
		SecureKeyID: root.SecureKeyID,
		Version:     root.Version,
		Kind:        derivedKindParams(parsed),
	}
}

func parseRootKeyOrWriteInvalid(w http.ResponseWriter, requestID string, generate func() string, root keyParams) (yggkey.ParsedKey, bool) {
	parsed, err := yggkey.Parse(root.KeyID)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, invalidEnvelopeWithID(requestID, generate, err.Error()))
		return yggkey.ParsedKey{}, false
	}
	return parsed, true
}

func writeForcedStatusEnvelope(w http.ResponseWriter, requestID string, generate func() string, secureKeyID string) bool {
	forced, ok := forcedStatusFromSecureKeyID(secureKeyID)
	if !ok {
		return false
	}
	writeJSON(w, forced.httpStatus, responseEnvelope[map[string]any]{
		ID:      responseIDWithGenerator(requestID, generate),
		Status:  forced.status,
		Message: forced.message,
		Data:    map[string]any{},
	})
	return true
}

func statusForDecodeError(err error) int {
	if errors.Is(err, errPayloadTooLarge) {
		return http.StatusRequestEntityTooLarge
	}
	return http.StatusBadRequest
}

func messageForDecodeError(err error) string {
	if errors.Is(err, errPayloadTooLarge) {
		return "payload too large"
	}
	return "invalid JSON payload"
}

// Mock-only forcing hook. Production integrity verification is intentionally out of scope here.
func forcedStatusFromSecureKeyID(secureKeyID string) (forcedStatus, bool) {
	switch secureKeyID {
	case "invalid":
		return forcedStatus{
			httpStatus: http.StatusBadRequest,
			status:     "invalid",
			message:    "forced by mock secureKeyId",
		}, true
	case "unauthorised":
		return forcedStatus{
			httpStatus: http.StatusUnauthorized,
			status:     "unauthorised",
			message:    "forced by mock secureKeyId",
		}, true
	case "outdated":
		return forcedStatus{
			httpStatus: http.StatusConflict,
			status:     "outdated",
			message:    "forced by mock secureKeyId",
		}, true
	default:
		return forcedStatus{}, false
	}
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(value); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(buf.Bytes())
}

func isDescendant(rootKey, childKey string) bool {
	return strings.HasPrefix(childKey, rootKey+":")
}

func responseIDWithGenerator(requestID string, generate func() string) string {
	if requestID != "" {
		return requestID
	}
	return generate()
}
