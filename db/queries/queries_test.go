package queries_test

import (
	"testing"

	"ptmuda.tech/backend/db/queries"
)

func TestQueriesLoad(t *testing.T) {
	names := queries.Names()
	if len(names) == 0 {
		t.Fatal("expected embedded queries to load, got 0")
	}

	expected := []string{
		"SchemaReady",
		"LoginAdminByIdentifier",
		"LoginAdminLock",
		"LoginAdminSaveSession",
		"LoginAdminFindSessionByAccess",
		"LoginAdminFindSessionByRefresh",
		"ProductCreate",
		"ProductGetByID",
		"ProductList",
		"CashflowCreate",
		"CashflowGetByID",
		"CashflowSummary",
		"DeviceGetByAPIKeyHash",
		"TelemetryIngest",
		"TelemetryLatest",
		"TelemetryListHistory",
		"TelemetryGetStats",
	}

	for _, name := range expected {
		q := queries.Get(name)
		if q == "" {
			t.Errorf("query %s is empty", name)
		}
	}
}
