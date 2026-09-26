package password

import (
	_ "embed"
	"strings"
	"sync"
)

// denylist.txt is embedded in the binary: the check never calls an external
// service. Source and license are in denylist.SOURCE.txt.
//
//go:embed denylist.txt
var denylistData string

// Denylist is a set of commonly used or compromised passwords.
type Denylist struct{ set map[string]struct{} }

// NewDenylist builds a list from lines, ignoring blank ones and comparing in
// lower case.
func NewDenylist(lines string) *Denylist {
	d := &Denylist{set: map[string]struct{}{}}
	for line := range strings.SplitSeq(lines, "\n") {
		if w := strings.ToLower(strings.TrimSpace(line)); w != "" {
			d.set[w] = struct{}{}
		}
	}
	return d
}

var (
	defaultOnce     sync.Once
	defaultDenylist *Denylist
)

// DefaultDenylist is the embedded list, parsed once.
func DefaultDenylist() *Denylist {
	defaultOnce.Do(func() { defaultDenylist = NewDenylist(denylistData) })
	return defaultDenylist
}

// Contains reports whether the password is in the list, ignoring letter case.
func (d *Denylist) Contains(plain string) bool {
	_, ok := d.set[strings.ToLower(plain)]
	return ok
}

// Len is the number of entries.
func (d *Denylist) Len() int { return len(d.set) }
