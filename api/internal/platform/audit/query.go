package audit

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"regexp"
	"time"

	"gorm.io/gorm"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/authz"
)

// Errors of the audit query; their messages are the API error codes.
var (
	ErrInvalidLimit  = errors.New("invalid_limit")
	ErrInvalidCursor = errors.New("invalid_cursor")
)

const (
	// ReadPermission is what the query requires.
	ReadPermission authz.Permission = "audit:log:read"

	defaultLimit = 50
	maxLimit     = 100
)

// Authorizer is authz.Authorizer: it decides by effective permissions and audits
// the denials.
type Authorizer interface {
	Require(ctx context.Context, p authz.Principal, perm authz.Permission) error
}

// Filter narrows a query; every set field must match (AND).
type Filter struct {
	EntityType  *string
	EntityID    *string
	ActorUserID *string
	Action      *string
	Outcome     *string
	From        *time.Time
	To          *time.Time
}

// Record is one stored entry, as the query returns it.
type Record struct {
	ID          string
	OccurredAt  time.Time
	ActorType   string
	ActorUserID *string
	Action      string
	EntityType  string
	EntityID    string
	Before      map[string]any
	After       map[string]any
	Outcome     string
	Reason      *string
	Context     map[string]any
	RequestID   *string
}

// Page is one page; NextCursor is empty on the last one.
type Page struct {
	Items      []Record
	NextCursor string
}

// Query reads the audit trail, newest first, with a keyset cursor over
// (occurred_at, id), so a page never repeats or skips an entry.
type Query struct {
	Authz Authorizer
	DB    *gorm.DB
}

var uuidText = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

type cursorWire struct {
	T  string `json:"t"`
	ID string `json:"id"`
}

func encodeCursor(at time.Time, id string) string {
	b, _ := json.Marshal(cursorWire{T: at.UTC().Format(time.RFC3339Nano), ID: id})
	return base64.RawURLEncoding.EncodeToString(b)
}

func decodeCursor(s string) (time.Time, string, error) {
	raw, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return time.Time{}, "", ErrInvalidCursor
	}
	var w cursorWire
	if err := json.Unmarshal(raw, &w); err != nil {
		return time.Time{}, "", ErrInvalidCursor
	}
	at, err := time.Parse(time.RFC3339Nano, w.T)
	if err != nil || !uuidText.MatchString(w.ID) {
		return time.Time{}, "", ErrInvalidCursor
	}
	return at, w.ID, nil
}

type row struct {
	ID          string
	OccurredAt  time.Time
	ActorType   string
	ActorUserID *string
	Action      string
	EntityType  string
	EntityID    string
	Before      *string
	After       *string
	Outcome     string
	Reason      *string
	Context     string
	RequestID   *string
}

// Search returns one page. It needs `audit:log:read`, checked before anything is read.
func (q *Query) Search(ctx context.Context, actor authz.Principal, f Filter, limit int, cursor string) (Page, error) {
	if q.Authz == nil || q.DB == nil {
		return Page{}, errors.New("auditoria: consulta mal configurada")
	}
	if err := q.Authz.Require(ctx, actor, ReadPermission); err != nil {
		return Page{}, err
	}
	switch {
	case limit == 0:
		limit = defaultLimit
	case limit < 0 || limit > maxLimit:
		return Page{}, ErrInvalidLimit
	}

	tx := q.DB.WithContext(ctx).Table("audit_log").Select(`id::text AS id, occurred_at, actor_type, actor_user_id::text AS actor_user_id,
		action, entity_type, entity_id, before::text AS before, after::text AS after, outcome, reason, context::text AS context, request_id`)
	if f.EntityType != nil {
		tx = tx.Where("entity_type = ?", *f.EntityType)
	}
	if f.EntityID != nil {
		tx = tx.Where("entity_id = ?", *f.EntityID)
	}
	if f.ActorUserID != nil {
		tx = tx.Where("actor_user_id = ?::uuid", *f.ActorUserID)
	}
	if f.Action != nil {
		tx = tx.Where("action = ?", *f.Action)
	}
	if f.Outcome != nil {
		tx = tx.Where("outcome = ?", *f.Outcome)
	}
	if f.From != nil {
		tx = tx.Where("occurred_at >= ?", *f.From)
	}
	if f.To != nil {
		tx = tx.Where("occurred_at <= ?", *f.To)
	}
	if cursor != "" {
		at, id, err := decodeCursor(cursor)
		if err != nil {
			return Page{}, err
		}
		tx = tx.Where("(occurred_at, id) < (?, ?::uuid)", at, id)
	}
	var rows []row
	if err := tx.Order("occurred_at DESC, id DESC").Limit(limit + 1).Scan(&rows).Error; err != nil {
		return Page{}, err
	}

	page := Page{}
	if len(rows) > limit {
		last := rows[limit-1]
		page.NextCursor = encodeCursor(last.OccurredAt, last.ID)
		rows = rows[:limit]
	}
	for _, r := range rows {
		rec, err := r.record()
		if err != nil {
			return Page{}, err
		}
		page.Items = append(page.Items, rec)
	}
	return page, nil
}

func (r row) record() (Record, error) {
	rec := Record{
		ID: r.ID, OccurredAt: r.OccurredAt, ActorType: r.ActorType, ActorUserID: r.ActorUserID, Action: r.Action,
		EntityType: r.EntityType, EntityID: r.EntityID, Outcome: r.Outcome, Reason: r.Reason, RequestID: r.RequestID,
	}
	var err error
	if rec.Before, err = objectOrNil(r.Before); err != nil {
		return Record{}, err
	}
	if rec.After, err = objectOrNil(r.After); err != nil {
		return Record{}, err
	}
	if err := json.Unmarshal([]byte(r.Context), &rec.Context); err != nil {
		return Record{}, err
	}
	if rec.Context == nil {
		rec.Context = map[string]any{}
	}
	return rec, nil
}

func objectOrNil(s *string) (map[string]any, error) {
	if s == nil {
		return nil, nil
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(*s), &m); err != nil {
		return nil, err
	}
	return m, nil
}
