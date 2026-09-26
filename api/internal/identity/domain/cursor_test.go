package domain

import (
	"testing"
	"time"
)

func TestListCursorRoundTripsAndKeepsTheInstantExactly(t *testing.T) {
	c := ListCursor{CreatedAt: time.Date(2026, 9, 26, 12, 30, 45, 123456000, time.UTC), ID: "0f8fad5b-d9cb-469f-a165-70867728950e"}

	got, err := DecodeCursor(EncodeCursor(c))

	if err != nil || !got.CreatedAt.Equal(c.CreatedAt) || got.ID != c.ID {
		t.Errorf("got = %+v, err = %v", got, err)
	}
}

func TestDecodeCursorRejectsGarbage(t *testing.T) {
	for _, s := range []string{"", "!!!", "e30", "eyJ0IjoiaG9qZSIsImlkIjoieCJ9", EncodeCursor(ListCursor{CreatedAt: time.Now(), ID: "não-é-uuid"})} {
		if got, err := DecodeCursor(s); err == nil {
			t.Errorf("DecodeCursor(%q) deveria falhar, veio %+v", s, got)
		}
	}
}
