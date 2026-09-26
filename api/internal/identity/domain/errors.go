package domain

import "errors"

// Errors of the identity use cases. Their messages are the API error codes.
var (
	ErrPrivilegeEscalation    = errors.New("privilege_escalation")
	ErrSelfChangeForbidden    = errors.New("self_change_forbidden")
	ErrLastAdmin              = errors.New("last_admin")
	ErrUserInactive           = errors.New("user_inactive")
	ErrAdminRoleRequired      = errors.New("admin_role_required")
	ErrInvalidCurrentPassword = errors.New("invalid_current_password")
	ErrPasswordUnchanged      = errors.New("password_unchanged")
	ErrInvalidLimit           = errors.New("invalid_limit")
	ErrInvalidCursor          = errors.New("invalid_cursor")
	ErrInvalidName            = errors.New("invalid_name")
	ErrInvalidEmail           = errors.New("invalid_email")
)
