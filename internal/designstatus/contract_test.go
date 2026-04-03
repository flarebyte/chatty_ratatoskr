package designstatus

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
)

type promiseStatus struct {
	Promise              string `json:"promise"`
	SourceSection        string `json:"source_section"`
	PlannedPhase         string `json:"planned_phase"`
	AcceptanceTest       string `json:"acceptance_test"`
	CurrentShippedStatus string `json:"current_shipped_status"`
}

func TestDesignPromiseStatus_MatchesPromptMatrix(t *testing.T) {
	repoRoot := filepath.Join("..", "..")

	artifactRows := readStatusArtifact(t, filepath.Join(repoRoot, "testdata", "design-promise-status.json"))
	promptRows := readPromptMatrix(t, filepath.Join(repoRoot, "scratch", "prompt.md"))

	if !slices.IsSortedFunc(artifactRows, func(a, b promiseStatus) int {
		return strings.Compare(strings.ToLower(a.Promise), strings.ToLower(b.Promise))
	}) {
		t.Fatal("design promise status artifact must be sorted by promise")
	}

	artifactByPromise := make(map[string]promiseStatus, len(artifactRows))
	for _, row := range artifactRows {
		if _, exists := artifactByPromise[row.Promise]; exists {
			t.Fatalf("duplicate promise in status artifact: %q", row.Promise)
		}
		if !isValidStatus(row.CurrentShippedStatus) {
			t.Fatalf("invalid shipped status %q for promise %q", row.CurrentShippedStatus, row.Promise)
		}
		artifactByPromise[row.Promise] = row
	}

	if len(promptRows) != len(artifactByPromise) {
		t.Fatalf("prompt matrix row count mismatch: prompt=%d artifact=%d", len(promptRows), len(artifactByPromise))
	}

	for _, promptRow := range promptRows {
		artifactRow, ok := artifactByPromise[promptRow.Promise]
		if !ok {
			t.Fatalf("missing promise in status artifact: %q", promptRow.Promise)
		}
		if promptRow != artifactRow {
			t.Fatalf("prompt matrix drift for %q:\nprompt=%+v\nartifact=%+v", promptRow.Promise, promptRow, artifactRow)
		}
		if promptRow.CurrentShippedStatus == "working" && !hasAcceptanceCoverage(repoRoot, promptRow.AcceptanceTest) {
			t.Fatalf("working promise %q has no machine-checked acceptance coverage: %q", promptRow.Promise, promptRow.AcceptanceTest)
		}
	}
}

func readStatusArtifact(t *testing.T, path string) []promiseStatus {
	t.Helper()

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read status artifact: %v", err)
	}

	var rows []promiseStatus
	if err := json.Unmarshal(raw, &rows); err != nil {
		t.Fatalf("parse status artifact: %v", err)
	}
	for i := range rows {
		rows[i] = normalizePromiseStatus(rows[i])
	}
	return rows
}

func readPromptMatrix(t *testing.T, path string) []promiseStatus {
	t.Helper()

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read prompt matrix: %v", err)
	}

	lines := strings.Split(string(raw), "\n")
	const header = "## Design Promise Matrix"

	inTable := false
	rows := make([]promiseStatus, 0)
	for _, line := range lines {
		if line == header {
			inTable = true
			continue
		}
		if !inTable {
			continue
		}
		if strings.HasPrefix(line, "Intentional narrowing:") {
			break
		}
		if !strings.HasPrefix(line, "|") {
			continue
		}
		columns := splitMarkdownRow(line)
		if len(columns) != 5 || columns[0] == "promise" || strings.HasPrefix(columns[0], "---") {
			continue
		}
		rows = append(rows, promiseStatus{
			Promise:              columns[0],
			SourceSection:        columns[1],
			PlannedPhase:         columns[2],
			AcceptanceTest:       columns[3],
			CurrentShippedStatus: columns[4],
		})
	}

	if len(rows) == 0 {
		t.Fatal("no design promise rows found in prompt matrix")
	}
	for i := range rows {
		rows[i] = normalizePromiseStatus(rows[i])
	}
	return rows
}

func splitMarkdownRow(line string) []string {
	trimmed := strings.TrimSpace(line)
	trimmed = strings.TrimPrefix(trimmed, "| ")
	trimmed = strings.TrimSuffix(trimmed, " |")
	parts := strings.Split(trimmed, " | ")
	columns := make([]string, 0, len(parts))
	for _, part := range parts {
		columns = append(columns, strings.TrimSpace(part))
	}
	return columns
}

func normalizePromiseStatus(row promiseStatus) promiseStatus {
	row.Promise = normalizeText(row.Promise)
	row.SourceSection = normalizeText(row.SourceSection)
	row.PlannedPhase = normalizeText(row.PlannedPhase)
	row.AcceptanceTest = normalizeText(row.AcceptanceTest)
	row.CurrentShippedStatus = normalizeText(row.CurrentShippedStatus)
	return row
}

func normalizeText(text string) string {
	return strings.TrimSpace(strings.ReplaceAll(text, "`", ""))
}

func isValidStatus(status string) bool {
	switch status {
	case "not started", "partial", "working", "intentionally deferred":
		return true
	default:
		return false
	}
}

func hasAcceptanceCoverage(repoRoot string, acceptance string) bool {
	if path := firstTSPath(acceptance); path != "" {
		if _, err := os.Stat(filepath.Join(repoRoot, path)); err == nil {
			return true
		}
	}

	testNames := testNamePattern.FindAllString(acceptance, -1)
	if len(testNames) == 0 {
		return false
	}

	for _, testName := range testNames {
		if !goTestExists(repoRoot, testName) {
			return false
		}
	}
	return true
}

func firstTSPath(text string) string {
	for _, field := range strings.Fields(text) {
		candidate := strings.Trim(field, "`',.()")
		if strings.HasSuffix(candidate, ".test.ts") {
			return candidate
		}
	}
	return ""
}

var testNamePattern = regexp.MustCompile(`Test[A-Za-z0-9_]+`)

func goTestExists(repoRoot, testName string) bool {
	pattern := "func " + testName + "("
	found := false
	_ = filepath.WalkDir(repoRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil || found {
			return nil
		}
		if d.IsDir() {
			switch d.Name() {
			case ".git", ".gocache", ".gomodcache", ".golangci-lint-cache", "node_modules":
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, "_test.go") {
			return nil
		}
		raw, readErr := os.ReadFile(path)
		if readErr == nil && strings.Contains(string(raw), pattern) {
			found = true
		}
		return nil
	})
	return found
}
