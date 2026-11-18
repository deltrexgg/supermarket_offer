package tenant

import (
	"strings"
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/deltrexgg/supermarket_offer/api/internal/models"
)

// dryRun builds a GORM handle that only writes SQL and never connects.
func dryRun(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(
		postgres.New(postgres.Config{DSN: "host=127.0.0.1 user=x dbname=x sslmode=disable"}),
		&gorm.Config{DryRun: true, DisableAutomaticPing: true},
	)
	if err != nil {
		t.Skipf("cannot build a dry-run handle: %v", err)
	}
	return db
}

func TestScopeFiltersBySupermarket(t *testing.T) {
	id := uint(7)
	var rows []models.Offer
	stmt := dryRun(t).Scopes(Scope(&id)).Find(&rows).Statement
	if !strings.Contains(stmt.SQL.String(), "supermarket_id = $1") {
		t.Fatalf("SQL has no tenant filter: %s", stmt.SQL.String())
	}
	if len(stmt.Vars) != 1 || stmt.Vars[0] != uint(7) {
		t.Fatalf("filter value = %v, want [7]", stmt.Vars)
	}
}

func TestScopeWithoutIDMatchesNothing(t *testing.T) {
	var rows []models.Offer
	stmt := dryRun(t).Scopes(Scope(nil)).Find(&rows).Statement
	if !strings.Contains(stmt.SQL.String(), "1 = 0") {
		t.Fatalf("a missing id must match nothing, SQL: %s", stmt.SQL.String())
	}
}
