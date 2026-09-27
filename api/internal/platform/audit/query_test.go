package audit

import (
	"strings"
	"testing"
	"time"
)

// AUD-01.3: occurred_at is always UTC in the response, whatever time zone the
// driver decoded it with (PostgreSQL's timestamptz follows the Go runtime's
// local zone, not necessarily UTC).
func TestRowRecordNormalizesOccurredAtToUTC(t *testing.T) {
	brt := time.FixedZone("BRT", -3*3600)
	local := time.Date(2026, 9, 26, 9, 0, 0, 0, brt) // 12:00 UTC

	rec, err := row{OccurredAt: local, Context: "{}"}.record()

	if err != nil {
		t.Fatal(err)
	}
	if rec.OccurredAt.Location() != time.UTC {
		t.Fatalf("Location() = %v, esperado UTC", rec.OccurredAt.Location())
	}
	if !rec.OccurredAt.Equal(local) {
		t.Errorf("o instante não pode mudar, só a representação: %v x %v", rec.OccurredAt, local)
	}
	body, _ := rec.OccurredAt.MarshalJSON()
	if !strings.HasSuffix(string(body), `Z"`) {
		t.Errorf("occurred_at = %s, deveria terminar em Z", body)
	}
}
