// purpose: Provide the deterministic in-memory state store used by snapshot and node workflows throughout the mock server.
// responsibilities:
// - Store and replace root snapshots.
// - Upsert and find individual entries under a root.
// - Return stable sorted snapshots and clear all state for admin reset flows.
// architecture_notes:
// - Ordering is canonicalized here so handlers do not each reimplement sorting rules.
// - The store is intentionally in-memory and mutex-backed; durability and external databases are out of scope.
// - Snapshot and node writes share this store to keep the mock server vertical slice practical and small.
package snapshot

import (
	"encoding/json"
	"sort"
	"sync"
)

type Store interface {
	Replace(rootKey Key, entries []KeyValue)
	Upsert(rootKey Key, entry KeyValue)
	Get(rootKey Key) Snapshot
	Find(rootKey Key, keyID string) (KeyValue, bool)
	Clear()
}

type Key struct {
	KeyID       string `json:"keyId"`
	SecureKeyID string `json:"secureKeyId,omitempty"`
	Version     string `json:"version,omitempty"`
}

type KeyValue struct {
	Key   Key    `json:"key"`
	Value string `json:"value,omitempty"`
}

type Snapshot struct {
	Key          Key        `json:"key"`
	KeyValueList []KeyValue `json:"keyValueList"`
}

type InMemoryStore struct {
	mu        sync.RWMutex
	snapshots map[string]Snapshot
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		snapshots: make(map[string]Snapshot),
	}
}

func (s *InMemoryStore) Replace(rootKey Key, entries []KeyValue) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.snapshots[rootKey.KeyID] = Snapshot{
		Key:          rootKey,
		KeyValueList: sortedEntries(entries),
	}
}

func (s *InMemoryStore) Upsert(rootKey Key, entry KeyValue) {
	s.mu.Lock()
	defer s.mu.Unlock()

	current, ok := s.snapshots[rootKey.KeyID]
	if !ok {
		s.snapshots[rootKey.KeyID] = Snapshot{
			Key:          rootKey,
			KeyValueList: []KeyValue{entry},
		}
		return
	}

	next := append([]KeyValue(nil), current.KeyValueList...)
	replaced := false
	for i := range next {
		if next[i].Key.KeyID == entry.Key.KeyID {
			next[i] = entry
			replaced = true
			break
		}
	}
	if !replaced {
		next = append(next, entry)
	}

	s.snapshots[rootKey.KeyID] = Snapshot{
		Key:          rootKey,
		KeyValueList: sortedEntries(next),
	}
}

func (s *InMemoryStore) Get(rootKey Key) Snapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()

	snapshot, ok := s.snapshots[rootKey.KeyID]
	if !ok {
		return Snapshot{
			Key:          rootKey,
			KeyValueList: []KeyValue{},
		}
	}
	return Snapshot{
		Key:          snapshot.Key,
		KeyValueList: append([]KeyValue(nil), snapshot.KeyValueList...),
	}
}

func (s *InMemoryStore) Find(rootKey Key, keyID string) (KeyValue, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	snapshot, ok := s.snapshots[rootKey.KeyID]
	if !ok {
		return KeyValue{}, false
	}
	for _, entry := range snapshot.KeyValueList {
		if entry.Key.KeyID == keyID {
			return entry, true
		}
	}
	return KeyValue{}, false
}

func (s *InMemoryStore) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.snapshots = make(map[string]Snapshot)
}

func MustJSON(snapshot Snapshot) string {
	data, err := json.Marshal(snapshot)
	if err != nil {
		panic(err)
	}
	return string(data)
}

func sortedEntries(entries []KeyValue) []KeyValue {
	out := append([]KeyValue(nil), entries...)
	sort.Slice(out, func(i, j int) bool {
		if out[i].Key.KeyID != out[j].Key.KeyID {
			return out[i].Key.KeyID < out[j].Key.KeyID
		}
		if out[i].Key.Version != out[j].Key.Version {
			return out[i].Key.Version < out[j].Key.Version
		}
		return out[i].Value < out[j].Value
	})
	return out
}
