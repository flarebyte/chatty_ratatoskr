package snapshot

import "testing"

func TestSnapshotStore_DeterministicOrdering(t *testing.T) {
	store := NewInMemoryStore()
	root := Key{
		KeyID:       "tenant:t8f3a1c2:group:g4b7d9e1:dashboard:d1e52f07",
		SecureKeyID: "ok",
	}

	entries := []KeyValue{
		{
			Key: Key{
				KeyID:       "tenant:t8f3a1c2:group:g4b7d9e1:dashboard:d1e52f07:note:n7c401c2:like:count",
				SecureKeyID: "ok",
				Version:     "v1",
			},
			Value: "3",
		},
		{
			Key: Key{
				KeyID:       "tenant:t8f3a1c2:group:g4b7d9e1:dashboard:d1e52f07:note:n7c401c2:text",
				SecureKeyID: "ok",
				Version:     "v1",
			},
			Value: "hello world",
		},
	}

	store.Replace(root, entries)
	got := store.Get(root)

	if got.Key.KeyID != root.KeyID {
		t.Fatalf("root key mismatch: got %q want %q", got.Key.KeyID, root.KeyID)
	}
	if len(got.KeyValueList) != 2 {
		t.Fatalf("unexpected entry count: got %d want 2", len(got.KeyValueList))
	}
	if got.KeyValueList[0].Key.KeyID != entries[0].Key.KeyID {
		t.Fatalf("first entry not sorted by key id: got %q", got.KeyValueList[0].Key.KeyID)
	}
	if got.KeyValueList[1].Key.KeyID != entries[1].Key.KeyID {
		t.Fatalf("second entry not sorted by key id: got %q", got.KeyValueList[1].Key.KeyID)
	}
}

func TestSnapshotStore_ReplaceAndGet(t *testing.T) {
	store := NewInMemoryStore()
	root := Key{
		KeyID:       "tenant:t8f3a1c2:group:g4b7d9e1:dashboard:d1e52f07",
		SecureKeyID: "ok",
	}

	store.Replace(root, []KeyValue{
		{
			Key: Key{
				KeyID:       root.KeyID + ":note:n1:text",
				SecureKeyID: "ok",
				Version:     "v1",
			},
			Value: "first",
		},
	})

	store.Replace(root, []KeyValue{
		{
			Key: Key{
				KeyID:       root.KeyID + ":note:n2:text",
				SecureKeyID: "ok",
				Version:     "v2",
			},
			Value: "second",
		},
	})

	got := store.Get(root)
	if len(got.KeyValueList) != 1 {
		t.Fatalf("replace should overwrite previous snapshot, got %d entries", len(got.KeyValueList))
	}
	if got.KeyValueList[0].Value != "second" {
		t.Fatalf("replace did not update snapshot value: got %q", got.KeyValueList[0].Value)
	}
}

func TestSnapshotStore_MissingSnapshotReturnsEmptyList(t *testing.T) {
	store := NewInMemoryStore()
	root := Key{
		KeyID:       "tenant:t8f3a1c2:group:g4b7d9e1:dashboard:d1e52f07",
		SecureKeyID: "ok",
	}

	got := store.Get(root)
	if got.Key.KeyID != root.KeyID {
		t.Fatalf("root key mismatch: got %q want %q", got.Key.KeyID, root.KeyID)
	}
	if len(got.KeyValueList) != 0 {
		t.Fatalf("missing snapshot should return empty list, got %d entries", len(got.KeyValueList))
	}
}

func TestSnapshotStore_DeterministicMarshal(t *testing.T) {
	store := NewInMemoryStore()
	root := Key{
		KeyID:       "tenant:t8f3a1c2:group:g4b7d9e1:dashboard:d1e52f07",
		SecureKeyID: "ok",
	}

	entries := []KeyValue{
		{
			Key: Key{
				KeyID:       root.KeyID + ":note:n7c401c2:like:count",
				SecureKeyID: "ok",
				Version:     "v1",
			},
			Value: "3",
		},
		{
			Key: Key{
				KeyID:       root.KeyID + ":note:n7c401c2:text",
				SecureKeyID: "ok",
				Version:     "v1",
			},
			Value: "hello world",
		},
	}

	store.Replace(root, entries)
	first := MustJSON(store.Get(root))
	second := MustJSON(store.Get(root))
	if first != second {
		t.Fatalf("marshal mismatch:\nfirst=%s\nsecond=%s", first, second)
	}
}
