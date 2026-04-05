package yggkey

import "testing"

func TestKeyParser_AcceptsCriticalFixtures(t *testing.T) {
	cases := []struct {
		name         string
		key          string
		wantTerminal string
		wantKinds    []string
	}{
		{
			name:         "group-root",
			key:          "tenant:t8f3a1c2:group:g4b7d9e1:dashboard:d1e52f07",
			wantTerminal: "id",
			wantKinds:    []string{"tenant", "group", "dashboard"},
		},
		{
			name:         "principal-note-text",
			key:          "tenant:t8f3a1c2:group:g4b7d9e1:user:u42c91ab:dashboard:d1e52f07:note:n7c401c2:text",
			wantTerminal: "leaf",
			wantKinds:    []string{"tenant", "group", "user", "dashboard", "note", "text"},
		},
		{
			name:         "principal-comment-text",
			key:          "tenant:t8f3a1c2:group:g4b7d9e1:user:u42c91ab:dashboard:d1e52f07:note:n7c401c2:comment:c38dd201:text",
			wantTerminal: "leaf",
			wantKinds:    []string{"tenant", "group", "user", "dashboard", "note", "comment", "text"},
		},
		{
			name:         "like-current-principal",
			key:          "tenant:t8f3a1c2:group:g4b7d9e1:dashboard:d1e52f07:note:n7c401c2:like:user:_",
			wantTerminal: "id",
			wantKinds:    []string{"tenant", "group", "dashboard", "note", "like", "user"},
		},
		{
			name:         "like-count",
			key:          "tenant:t8f3a1c2:group:g4b7d9e1:dashboard:d1e52f07:note:n7c401c2:like:count",
			wantTerminal: "derived",
			wantKinds:    []string{"tenant", "group", "dashboard", "note", "like", "count"},
		},
		{
			name:         "language-current-principal",
			key:          "department:d9a8c102:region:r6be41f0:member:m17aa9f3:dashboard:d1e52f07:note:n7c401c2:language:_",
			wantTerminal: "branch",
			wantKinds:    []string{"department", "region", "member", "dashboard", "note", "language"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Parse(tc.key)
			if err != nil {
				t.Fatalf("Parse(%q) error = %v", tc.key, err)
			}
			if got.Canonical != tc.key {
				t.Fatalf("Canonical mismatch: got %q want %q", got.Canonical, tc.key)
			}
			if got.TerminalKind != tc.wantTerminal {
				t.Fatalf("TerminalKind mismatch: got %q want %q", got.TerminalKind, tc.wantTerminal)
			}
			if len(got.KindPath) != len(tc.wantKinds) {
				t.Fatalf("KindPath length mismatch: got %v want %v", got.KindPath, tc.wantKinds)
			}
			for i, want := range tc.wantKinds {
				if got.KindPath[i] != want {
					t.Fatalf("KindPath[%d] mismatch: got %q want %q", i, got.KindPath[i], want)
				}
			}
		})
	}
}

func TestKeyParser_RejectsMalformedExamples(t *testing.T) {
	cases := []struct {
		name string
		key  string
	}{
		{name: "missing-root-id", key: "tenant:t8f3a1c2:group:g4b7d9e1:dashboard"},
		{name: "unknown-level1-label", key: "workspace:w1234567:group:g4b7d9e1:dashboard:d1e52f07"},
		{name: "principal-placeholder-without-label", key: "tenant:t8f3a1c2:group:g4b7d9e1:_:dashboard:d1e52f07"},
		{name: "derived-leaf-with-child", key: "tenant:t8f3a1c2:group:g4b7d9e1:dashboard:d1e52f07:note:n7c401c2:like:count:text"},
		{name: "leaf-with-child", key: "tenant:t8f3a1c2:group:g4b7d9e1:user:u42c91ab:dashboard:d1e52f07:note:n7c401c2:text:comment:c38dd201"},
		{name: "unknown-derived-leaf", key: "tenant:t8f3a1c2:group:g4b7d9e1:dashboard:d1e52f07:note:n7c401c2:like:total"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := Parse(tc.key); err == nil {
				t.Fatalf("Parse(%q) succeeded, want error", tc.key)
			}
		})
	}
}

func TestKeyParser_DeterministicMarshal(t *testing.T) {
	key := "tenant:t8f3a1c2:group:g4b7d9e1:user:u42c91ab:dashboard:d1e52f07:note:n7c401c2:text"

	first, err := Parse(key)
	if err != nil {
		t.Fatalf("Parse first: %v", err)
	}
	second, err := Parse(key)
	if err != nil {
		t.Fatalf("Parse second: %v", err)
	}

	gotFirst := MustJSON(first)
	gotSecond := MustJSON(second)
	if gotFirst != gotSecond {
		t.Fatalf("marshal mismatch:\nfirst=%s\nsecond=%s", gotFirst, gotSecond)
	}
}

func TestKey_DerivedKind(t *testing.T) {
	cases := []struct {
		name string
		key  string
		want []string
	}{
		{
			name: "root only",
			key:  "tenant:t8f3a1c2:group:g4b7d9e1:dashboard:d1e52f07",
			want: []string{"dashboard"},
		},
		{
			name: "note text",
			key:  "tenant:t8f3a1c2:group:g4b7d9e1:user:u42c91ab:dashboard:d1e52f07:note:n7c401c2:text",
			want: []string{"dashboard", "note", "text"},
		},
		{
			name: "like count",
			key:  "tenant:t8f3a1c2:group:g4b7d9e1:dashboard:d1e52f07:note:n7c401c2:like:count",
			want: []string{"dashboard", "note", "like", "count"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			parsed, err := Parse(tc.key)
			if err != nil {
				t.Fatalf("Parse(%q) error = %v", tc.key, err)
			}
			got := parsed.DerivedKind()
			if len(got.Hierarchy) != len(tc.want) {
				t.Fatalf("Hierarchy length mismatch: got %v want %v", got.Hierarchy, tc.want)
			}
			for i, want := range tc.want {
				if got.Hierarchy[i] != want {
					t.Fatalf("Hierarchy[%d] mismatch: got %q want %q", i, got.Hierarchy[i], want)
				}
			}
		})
	}
}

func TestKey_AcceptanceExamples(t *testing.T) {
	TestKeyParser_AcceptsCriticalFixtures(t)
}

func TestKey_RejectionExamples(t *testing.T) {
	TestKeyParser_RejectsMalformedExamples(t)
}
