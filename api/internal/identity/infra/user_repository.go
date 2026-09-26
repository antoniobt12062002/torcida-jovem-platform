package infra

import (
	"context"
	"database/sql/driver"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/domain"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/authz"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/database"
)

var uuidFormat = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// bytesParam makes GORM pass a []byte as one bound value; without it, GORM
// expands a byte slice into a list of numbers inside "VALUES (?".
type bytesParam []byte

func (b bytesParam) Value() (driver.Value, error) { return []byte(b), nil }

// conn is the transaction of the use case when there is one, else the pool.
func conn(ctx context.Context, db *gorm.DB) *gorm.DB {
	if tx, ok := database.TxFrom(ctx); ok {
		return tx.WithContext(ctx)
	}
	return db.WithContext(ctx)
}

// UserRepository persists users and their role assignments.
type UserRepository struct{ db *gorm.DB }

func NewUserRepository(db *gorm.DB) *UserRepository { return &UserRepository{db: db} }

type userRow struct {
	ID                 string
	Email              string
	Name               string
	PasswordHash       string
	IsActive           bool
	MustChangePassword bool
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

func (r userRow) user() domain.User {
	return domain.User{
		ID: r.ID, Email: r.Email, Name: r.Name, PasswordHash: r.PasswordHash, Active: r.IsActive,
		MustChangePassword: r.MustChangePassword, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt,
	}
}

const userColumns = `id::text AS id, email, name, password_hash, is_active, must_change_password, created_at, updated_at`

// Create inserts the user. The e-mail must already be normalized. A repeated
// e-mail, ignoring case, returns domain.ErrEmailTaken; with concurrent creations
// of the same e-mail exactly one wins, because the database decides.
func (r *UserRepository) Create(ctx context.Context, u domain.User) (domain.User, error) {
	if norm, err := domain.NormalizeEmail(u.Email); err != nil || norm != u.Email {
		return domain.User{}, errors.New("usuário: o e-mail deve estar normalizado")
	}
	var row userRow
	err := conn(ctx, r.db).Raw(`INSERT INTO users (email, name, password_hash, is_active, must_change_password)
		VALUES (?, ?, ?, ?, ?) RETURNING `+userColumns,
		u.Email, u.Name, u.PasswordHash, u.Active, u.MustChangePassword).Scan(&row).Error
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok && pgErr.Code == "23505" && pgErr.ConstraintName == "users_email_uq" {
			return domain.User{}, domain.ErrEmailTaken
		}
		return domain.User{}, fmt.Errorf("usuário: criar: %w", err)
	}
	return row.user(), nil
}

// FindByEmail looks the user up by e-mail, ignoring case and surrounding spaces.
// An invalid e-mail is just a user that does not exist.
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	norm, err := domain.NormalizeEmail(email)
	if err != nil {
		return domain.User{}, domain.ErrUserNotFound
	}
	return r.find(ctx, "lower(email) = ?", norm)
}

// FindByID looks the user up by id; a malformed id is a user that does not exist.
func (r *UserRepository) FindByID(ctx context.Context, id string) (domain.User, error) {
	if !uuidFormat.MatchString(id) {
		return domain.User{}, domain.ErrUserNotFound
	}
	return r.find(ctx, "id = ?::uuid", id)
}

func (r *UserRepository) find(ctx context.Context, where string, arg any) (domain.User, error) {
	var rows []userRow
	if err := conn(ctx, r.db).Raw("SELECT "+userColumns+" FROM users WHERE "+where, arg).Scan(&rows).Error; err != nil {
		return domain.User{}, fmt.Errorf("usuário: consultar: %w", err)
	}
	if len(rows) == 0 {
		return domain.User{}, domain.ErrUserNotFound
	}
	return rows[0].user(), nil
}

// SetPassword replaces the password hash and the must-change flag.
func (r *UserRepository) SetPassword(ctx context.Context, id, hash string, mustChange bool) error {
	return r.update(ctx, id, "password_hash = ?, must_change_password = ?", hash, mustChange)
}

// SetMustChangePassword sets only the must-change flag.
func (r *UserRepository) SetMustChangePassword(ctx context.Context, id string, mustChange bool) error {
	return r.update(ctx, id, "must_change_password = ?", mustChange)
}

// SetActive activates or deactivates the user.
func (r *UserRepository) SetActive(ctx context.Context, id string, active bool) error {
	return r.update(ctx, id, "is_active = ?", active)
}

func (r *UserRepository) update(ctx context.Context, id, set string, args ...any) error {
	if !uuidFormat.MatchString(id) {
		return domain.ErrUserNotFound
	}
	res := conn(ctx, r.db).Exec("UPDATE users SET "+set+", updated_at = now() WHERE id = ?::uuid", append(args, id)...)
	if res.Error != nil {
		return fmt.Errorf("usuário: atualizar: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}

// SetRoles replaces the roles of the user, in one unit of work. The role set
// cannot be empty and every role must exist (be synchronized).
func (r *UserRepository) SetRoles(ctx context.Context, userID string, roles []domain.Role) error {
	if len(roles) == 0 {
		return errors.New("usuário: o conjunto de papéis não pode ser vazio")
	}
	if !uuidFormat.MatchString(userID) {
		return domain.ErrUserNotFound
	}
	return database.WithTx(ctx, r.db, func(ctx context.Context) error {
		tx, _ := database.TxFrom(ctx)
		if err := tx.Exec("DELETE FROM user_roles WHERE user_id = ?::uuid", userID).Error; err != nil {
			return fmt.Errorf("usuário: limpar papéis: %w", err)
		}
		for _, role := range roles {
			res := tx.Exec(`INSERT INTO user_roles (user_id, role_id)
				SELECT ?::uuid, id FROM roles WHERE name = ? ON CONFLICT DO NOTHING`, userID, string(role))
			if res.Error != nil {
				return fmt.Errorf("usuário: atribuir papel: %w", res.Error)
			}
			if res.RowsAffected == 0 && !roleAlreadyAssigned(tx, userID, role) {
				return fmt.Errorf("usuário: papel %q não existe", role)
			}
		}
		return nil
	})
}

func roleAlreadyAssigned(tx *gorm.DB, userID string, role domain.Role) bool {
	var n int64
	_ = tx.Raw(`SELECT count(*) FROM user_roles ur JOIN roles r ON r.id = ur.role_id WHERE ur.user_id = ?::uuid AND r.name = ?`,
		userID, string(role)).Scan(&n).Error
	return n > 0
}

// AdminSetLockKey is the transaction-level advisory lock that serializes every
// operation that can reduce who administers access (deactivating a user,
// withdrawing the membership, changing roles), so two of them running together
// cannot leave the system without an administrator.
const AdminSetLockKey int64 = 0x746a5f61646d696e // "tj_admin"

// LockAdminSet takes the admin-set lock for the current transaction. It fails
// outside a transaction, because the lock would end at once.
func (r *UserRepository) LockAdminSet(ctx context.Context) error {
	tx, ok := database.TxFrom(ctx)
	if !ok {
		return errors.New("usuário: LockAdminSet exige uma transação")
	}
	if err := tx.Exec("SELECT pg_advisory_xact_lock(?)", AdminSetLockKey).Error; err != nil {
		return fmt.Errorf("usuário: lock do conjunto de administradores: %w", err)
	}
	return nil
}

// ActiveHoldersOf counts the active users that hold the (active) permission
// through their roles, not counting the excluded user id (may be empty).
func (r *UserRepository) ActiveHoldersOf(ctx context.Context, perm authz.Permission, exclude string) (int, error) {
	if exclude != "" && !uuidFormat.MatchString(exclude) {
		return 0, domain.ErrUserNotFound
	}
	var n int64
	q := `SELECT count(DISTINCT u.id) FROM users u
		JOIN user_roles ur ON ur.user_id = u.id
		JOIN role_permissions rp ON rp.role_id = ur.role_id
		JOIN permissions p ON p.id = rp.permission_id
		WHERE u.is_active AND p.is_active AND p.name = ?`
	args := []any{string(perm)}
	if exclude != "" {
		q += " AND u.id <> ?::uuid"
		args = append(args, exclude)
	}
	if err := conn(ctx, r.db).Raw(q, args...).Scan(&n).Error; err != nil {
		return 0, fmt.Errorf("usuário: contar titulares: %w", err)
	}
	return int(n), nil
}

// List returns up to limit users, newest first, after the cursor (nil starts at
// the newest). Each item carries the roles and the summary of the active
// administrative membership; the password hash is never read.
func (r *UserRepository) List(ctx context.Context, filter domain.ListFilter, after *domain.ListCursor, limit int) ([]domain.UserSummary, error) {
	q := `SELECT id::text AS id, email, name, is_active, must_change_password, created_at FROM users u WHERE true`
	var args []any
	if filter.Active != nil {
		q += " AND u.is_active = ?"
		args = append(args, *filter.Active)
	}
	if filter.Role != nil {
		q += " AND EXISTS (SELECT 1 FROM user_roles ur JOIN roles r ON r.id = ur.role_id WHERE ur.user_id = u.id AND r.name = ?)"
		args = append(args, string(*filter.Role))
	}
	if after != nil {
		q += " AND (u.created_at, u.id) < (?, ?::uuid)"
		args = append(args, after.CreatedAt, after.ID)
	}
	q += " ORDER BY u.created_at DESC, u.id DESC LIMIT ?"
	args = append(args, limit)

	var rows []struct {
		ID                 string
		Email              string
		Name               string
		IsActive           bool
		MustChangePassword bool
		CreatedAt          time.Time
	}
	c := conn(ctx, r.db)
	if err := c.Raw(q, args...).Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("usuários: listar: %w", err)
	}
	out := make([]domain.UserSummary, len(rows))
	ids := make([]string, len(rows))
	index := map[string]int{}
	for i, row := range rows {
		out[i] = domain.UserSummary{ID: row.ID, Email: row.Email, Name: row.Name, Active: row.IsActive, MustChangePassword: row.MustChangePassword, CreatedAt: row.CreatedAt}
		ids[i] = row.ID
		index[row.ID] = i
	}
	if len(ids) == 0 {
		return out, nil
	}

	var roleRows []struct{ UserID, Name string }
	if err := c.Raw(`SELECT ur.user_id::text AS user_id, r.name FROM user_roles ur JOIN roles r ON r.id = ur.role_id
		WHERE ur.user_id = ANY(?::uuid[]) ORDER BY r.name`, pgUUIDArray(ids)).Scan(&roleRows).Error; err != nil {
		return nil, fmt.Errorf("usuários: papéis da página: %w", err)
	}
	for _, rr := range roleRows {
		out[index[rr.UserID]].Roles = append(out[index[rr.UserID]].Roles, domain.Role(rr.Name))
	}
	var memRows []struct {
		UserID    string
		Reason    string
		GrantedAt time.Time
	}
	if err := c.Raw(`SELECT user_id::text AS user_id, reason, granted_at FROM admin_memberships
		WHERE user_id = ANY(?::uuid[]) AND revoked_at IS NULL`, pgUUIDArray(ids)).Scan(&memRows).Error; err != nil {
		return nil, fmt.Errorf("usuários: vínculos da página: %w", err)
	}
	for _, mr := range memRows {
		out[index[mr.UserID]].AdminMembership = &domain.MembershipSummary{Reason: mr.Reason, GrantedAt: mr.GrantedAt}
	}
	return out, nil
}

// pgUUIDArray formats ids (already validated UUIDs from the database) as a
// PostgreSQL array literal.
func pgUUIDArray(ids []string) string { return "{" + strings.Join(ids, ",") + "}" }
