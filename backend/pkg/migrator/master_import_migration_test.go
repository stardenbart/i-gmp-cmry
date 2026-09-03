package migrator

import (
	"os"
	"strings"
	"testing"
)

func TestMasterImportMigrationCanBeSplitSafely(t *testing.T) {
	content, err := os.ReadFile("../../migrations/045_master_import_and_unique_constraints.sql")
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}
	statements := splitSQLStatements(string(content))
	if len(statements) < 15 {
		t.Fatalf("expected complete migration statements, got %d", len(statements))
	}
	for i, statement := range statements {
		if strings.Count(statement, "$$")%2 != 0 {
			t.Fatalf("statement %d has an unterminated dollar quote", i+1)
		}
	}
	joined := strings.Join(statements, "\n")
	for _, expected := range []string{
		"master_clean_text", "uq_aspek_area_normalized_name", "Master_Import_Batch", "PERM-MSTR-I",
	} {
		if !strings.Contains(joined, expected) {
			t.Fatalf("migration statement for %s is missing", expected)
		}
	}
}
