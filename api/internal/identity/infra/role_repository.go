// Package infra holds the identity persistence.
package infra

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"gorm.io/gorm"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/domain"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/audit"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/authz"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/database"
)

// syncLockKey is the transaction-level advisory lock that serializes the
// synchronization when several instances start together.
const syncLockKey int64 = 0x746a5f72626163 // "tj_rbac"

// RoleRepository persists roles, permissions and the role-permission links.
type RoleRepository struct {
	db  *gorm.DB
	rec *audit.Recorder
}

func NewRoleRepository(db *gorm.DB, rec *audit.Recorder) *RoleRepository {
	return &RoleRepository{db: db, rec: rec}
}

// SyncResult tells whether the synchronization changed anything.
type SyncResult struct{ Changed bool }

type permRow struct {
	Name        string
	Description string
	IsActive    bool
}

type linkKey struct{ Role, Permission string }

// state is the database as the synchronization sees it.
type state struct {
	permissions map[string]permRow
	roles       map[string]string // name -> description
	links       map[linkKey]struct{}
}

// Sync makes the database match the matrix, idempotently, in one transaction
// under an advisory lock. Permissions that left the code are marked inactive and
// never deleted; their links are dropped. When something changed it records
// `rbac.sync` (no actor) with the full state before and after and the diff, so
// the history can be rebuilt; with no change it records nothing. If the audit
// write fails the whole synchronization rolls back.
func (r *RoleRepository) Sync(ctx context.Context, m domain.Matrix) (SyncResult, error) {
	var res SyncResult
	err := database.WithTx(ctx, r.db, func(ctx context.Context) error {
		tx, _ := database.TxFrom(ctx)
		if err := tx.Exec("SELECT pg_advisory_xact_lock(?)", syncLockKey).Error; err != nil {
			return fmt.Errorf("rbac sync: lock: %w", err)
		}
		before, err := readState(tx)
		if err != nil {
			return err
		}
		diff := computeDiff(before, m)
		if diff.empty() {
			return nil
		}
		if err := apply(tx, m, diff); err != nil {
			return err
		}
		after, err := readState(tx)
		if err != nil {
			return err
		}
		res.Changed = true
		return r.rec.Record(ctx, audit.Entry{
			ActorType:  audit.ActorSystem,
			Action:     audit.RBACSync,
			EntityType: "rbac",
			EntityID:   "matrix",
			Outcome:    audit.OutcomeSuccess,
			Before:     before.snapshot(),
			After:      after.snapshot(),
			Context:    diff.context(),
		})
	})
	return res, err
}

func readState(tx *gorm.DB) (state, error) {
	s := state{permissions: map[string]permRow{}, roles: map[string]string{}, links: map[linkKey]struct{}{}}
	var perms []permRow
	if err := tx.Raw("SELECT name, description, is_active FROM permissions").Scan(&perms).Error; err != nil {
		return s, fmt.Errorf("rbac sync: ler permissões: %w", err)
	}
	for _, p := range perms {
		s.permissions[p.Name] = p
	}
	var roles []struct{ Name, Description string }
	if err := tx.Raw("SELECT name, description FROM roles").Scan(&roles).Error; err != nil {
		return s, fmt.Errorf("rbac sync: ler papéis: %w", err)
	}
	for _, ro := range roles {
		s.roles[ro.Name] = ro.Description
	}
	var links []linkKey
	if err := tx.Raw(`SELECT r.name AS role, p.name AS permission
		FROM role_permissions rp JOIN roles r ON r.id = rp.role_id JOIN permissions p ON p.id = rp.permission_id`).Scan(&links).Error; err != nil {
		return s, fmt.Errorf("rbac sync: ler vínculos: %w", err)
	}
	for _, l := range links {
		s.links[l] = struct{}{}
	}
	return s, nil
}

// snapshot is the JSON-friendly full state stored in the audit entry.
func (s state) snapshot() map[string]any {
	perms := make([]any, 0, len(s.permissions))
	for _, name := range sortedKeys(s.permissions) {
		perms = append(perms, map[string]any{"name": name, "active": s.permissions[name].IsActive})
	}
	roles := map[string]any{}
	for name := range s.roles {
		roles[name] = []any{}
	}
	perRole := map[string][]string{}
	for l := range s.links {
		perRole[l.Role] = append(perRole[l.Role], l.Permission)
	}
	for role, list := range perRole {
		slices.Sort(list)
		out := make([]any, len(list))
		for i, p := range list {
			out[i] = p
		}
		roles[role] = out
	}
	return map[string]any{"permissions": perms, "roles": roles}
}

type diff struct {
	permissionsAdded, permissionsDeactivated, permissionsReactivated, permissionsUpdated []string
	rolesAdded, rolesUpdated                                                             []string
	linksAdded, linksRemoved                                                             []linkKey
}

func (d diff) empty() bool {
	return len(d.permissionsAdded)+len(d.permissionsDeactivated)+len(d.permissionsReactivated)+len(d.permissionsUpdated)+
		len(d.rolesAdded)+len(d.rolesUpdated)+len(d.linksAdded)+len(d.linksRemoved) == 0
}

func (d diff) context() map[string]any {
	links := func(ls []linkKey) []any {
		out := make([]any, len(ls))
		for i, l := range ls {
			out[i] = l.Role + ":" + l.Permission
		}
		return out
	}
	strs := func(ss []string) []any {
		out := make([]any, len(ss))
		for i, s := range ss {
			out[i] = s
		}
		return out
	}
	return map[string]any{
		"permissions_added":       strs(d.permissionsAdded),
		"permissions_deactivated": strs(d.permissionsDeactivated),
		"permissions_reactivated": strs(d.permissionsReactivated),
		"permissions_updated":     strs(d.permissionsUpdated),
		"roles_added":             strs(d.rolesAdded),
		"roles_updated":           strs(d.rolesUpdated),
		"links_added":             links(d.linksAdded),
		"links_removed":           links(d.linksRemoved),
	}
}

func computeDiff(cur state, m domain.Matrix) diff {
	var d diff
	desired := map[string]authz.Definition{}
	for _, def := range m.Definitions {
		desired[string(def.Permission)] = def
		row, exists := cur.permissions[string(def.Permission)]
		switch {
		case !exists:
			d.permissionsAdded = append(d.permissionsAdded, string(def.Permission))
		case !row.IsActive:
			d.permissionsReactivated = append(d.permissionsReactivated, string(def.Permission))
		}
		if exists && row.Description != def.Description {
			d.permissionsUpdated = append(d.permissionsUpdated, string(def.Permission))
		}
	}
	for name, row := range cur.permissions {
		if _, ok := desired[name]; !ok && row.IsActive {
			d.permissionsDeactivated = append(d.permissionsDeactivated, name)
		}
	}
	for _, role := range domain.AllRoles {
		desc, exists := cur.roles[string(role)]
		if !exists {
			d.rolesAdded = append(d.rolesAdded, string(role))
		} else if desc != role.Description() {
			d.rolesUpdated = append(d.rolesUpdated, string(role))
		}
	}
	want := map[linkKey]struct{}{}
	for role, perms := range m.Grants {
		for _, p := range perms {
			want[linkKey{string(role), string(p)}] = struct{}{}
		}
	}
	for l := range want {
		if _, ok := cur.links[l]; !ok {
			d.linksAdded = append(d.linksAdded, l)
		}
	}
	for l := range cur.links {
		if _, ok := want[l]; !ok {
			d.linksRemoved = append(d.linksRemoved, l)
		}
	}
	for _, s := range [][]string{d.permissionsAdded, d.permissionsDeactivated, d.permissionsReactivated, d.permissionsUpdated, d.rolesAdded, d.rolesUpdated} {
		slices.Sort(s)
	}
	byKey := func(a, b linkKey) int { return strings.Compare(a.Role+":"+a.Permission, b.Role+":"+b.Permission) }
	slices.SortFunc(d.linksAdded, byKey)
	slices.SortFunc(d.linksRemoved, byKey)
	return d
}

func apply(tx *gorm.DB, m domain.Matrix, d diff) error {
	descriptions := map[string]string{}
	for _, def := range m.Definitions {
		descriptions[string(def.Permission)] = def.Description
	}
	for _, role := range append(append([]string{}, d.rolesAdded...), d.rolesUpdated...) {
		if err := tx.Exec(`INSERT INTO roles (name, description) VALUES (?, ?)
			ON CONFLICT (name) DO UPDATE SET description = EXCLUDED.description`, role, domain.Role(role).Description()).Error; err != nil {
			return fmt.Errorf("rbac sync: papel %s: %w", role, err)
		}
	}
	for _, name := range d.permissionsAdded {
		if err := tx.Exec("INSERT INTO permissions (name, description) VALUES (?, ?)", name, descriptions[name]).Error; err != nil {
			return fmt.Errorf("rbac sync: permissão %s: %w", name, err)
		}
	}
	for _, name := range d.permissionsReactivated {
		if err := tx.Exec("UPDATE permissions SET is_active = true, description = ? WHERE name = ?", descriptions[name], name).Error; err != nil {
			return fmt.Errorf("rbac sync: reativar %s: %w", name, err)
		}
	}
	for _, name := range d.permissionsUpdated {
		if err := tx.Exec("UPDATE permissions SET description = ? WHERE name = ?", descriptions[name], name).Error; err != nil {
			return fmt.Errorf("rbac sync: descrição de %s: %w", name, err)
		}
	}
	for _, name := range d.permissionsDeactivated {
		if err := tx.Exec("UPDATE permissions SET is_active = false WHERE name = ?", name).Error; err != nil {
			return fmt.Errorf("rbac sync: desativar %s: %w", name, err)
		}
	}
	for _, l := range d.linksRemoved {
		err := tx.Exec(`DELETE FROM role_permissions
			WHERE role_id = (SELECT id FROM roles WHERE name = ?) AND permission_id = (SELECT id FROM permissions WHERE name = ?)`,
			l.Role, l.Permission).Error
		if err != nil {
			return fmt.Errorf("rbac sync: remover vínculo %s: %w", l.Permission, err)
		}
	}
	for _, l := range d.linksAdded {
		err := tx.Exec(`INSERT INTO role_permissions (role_id, permission_id)
			SELECT r.id, p.id FROM roles r, permissions p WHERE r.name = ? AND p.name = ?`, l.Role, l.Permission).Error
		if err != nil {
			return fmt.Errorf("rbac sync: criar vínculo %s: %w", l.Permission, err)
		}
	}
	return nil
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys
}

// EffectivePermissions is the union of the permissions of every role of the
// user, counting only active permissions, sorted. A user with no roles, or that
// does not exist, has none.
func (r *RoleRepository) EffectivePermissions(ctx context.Context, userID string) ([]authz.Permission, error) {
	var names []string
	err := r.db.WithContext(ctx).Raw(`SELECT DISTINCT p.name
		FROM user_roles ur
		JOIN role_permissions rp ON rp.role_id = ur.role_id
		JOIN permissions p ON p.id = rp.permission_id
		WHERE ur.user_id = ? AND p.is_active
		ORDER BY p.name`, userID).Scan(&names).Error
	if err != nil {
		return nil, fmt.Errorf("permissões efetivas: %w", err)
	}
	out := make([]authz.Permission, len(names))
	for i, n := range names {
		out[i] = authz.Permission(n)
	}
	return out, nil
}

// RolesOf lists the roles of the user, sorted by name.
func (r *RoleRepository) RolesOf(ctx context.Context, userID string) ([]domain.Role, error) {
	var names []string
	err := r.db.WithContext(ctx).Raw(`SELECT ro.name FROM user_roles ur JOIN roles ro ON ro.id = ur.role_id
		WHERE ur.user_id = ? ORDER BY ro.name`, userID).Scan(&names).Error
	if err != nil {
		return nil, fmt.Errorf("papéis do usuário: %w", err)
	}
	out := make([]domain.Role, len(names))
	for i, n := range names {
		out[i] = domain.Role(n)
	}
	return out, nil
}
