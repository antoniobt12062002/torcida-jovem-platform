package infra

import (
	"context"
	"database/sql/driver"
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/domain"
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
