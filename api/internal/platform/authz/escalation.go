package authz

import "slices"

// Covers returns, sorted and without repetition, the permissions in needed that
// the principal does not hold. An empty result means the principal may grant,
// remove or change anything made of those permissions: nobody can hand out a
// permission they do not have. Only the effective permissions of the principal
// count.
func Covers(p Principal, needed []Permission) (missing []Permission) {
	for _, perm := range needed {
		if !p.Has(perm) && !slices.Contains(missing, perm) {
			missing = append(missing, perm)
		}
	}
	slices.Sort(missing)
	return missing
}
