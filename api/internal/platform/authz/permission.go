// Package authz holds permissions, the authenticated principal and the checks
// that decide access. Decisions use effective permissions only, never role names.
package authz

import (
	"fmt"
	"regexp"
	"strings"
)

// Permission is "<module>:<resource>:<action>", for example
// "financeiro:lancamento:read".
type Permission string

var permissionFormat = regexp.MustCompile(`^[a-z_]+:[a-z_]+:[a-z_]+$`)

// ParsePermission validates the form of a permission.
func ParsePermission(s string) (Permission, error) {
	if !permissionFormat.MatchString(s) {
		return "", fmt.Errorf("authz: permissão %q fora do formato módulo:recurso:ação", s)
	}
	return Permission(s), nil
}

// Action is the last segment, for example "read" or "approve".
func (p Permission) Action() string {
	return p[strings.LastIndexByte(string(p), ':')+1:].String()
}

func (p Permission) String() string { return string(p) }

// Definition declares a permission for the catalog.
type Definition struct {
	Permission  Permission
	Description string
	// CommonRead marks an ordinary read whose denial stays only in the access
	// log instead of the audit trail. Only permissions whose action is "read"
	// can be common reads: administration, security, RBAC, financial and
	// institutional permissions are always audited.
	CommonRead bool
}

// Validate checks the definition.
func (d Definition) Validate() error {
	if _, err := ParsePermission(string(d.Permission)); err != nil {
		return err
	}
	if d.CommonRead && d.Permission.Action() != "read" {
		return fmt.Errorf("authz: %s não pode ser leitura comum: só ações read", d.Permission)
	}
	return nil
}
