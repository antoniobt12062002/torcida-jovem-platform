package audit

import (
	"context"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/authz"
)

// DeniedHook returns the authz.Authorizer denial hook that records `authz.denied`
// as a security event: the user as actor and the required permission as target.
// Like every security event, a failure to write it is logged as an incident and
// never changes the 403 response.
func DeniedHook(rec *Recorder) authz.DeniedFunc {
	return func(ctx context.Context, p authz.Principal, perm authz.Permission) {
		rec.RecordSecurity(ctx, Entry{
			ActorType:  ActorUser,
			ActorID:    p.UserID,
			Action:     AuthzDenied,
			EntityType: "permission",
			EntityID:   string(perm),
			Outcome:    OutcomeDenied,
		})
	}
}
