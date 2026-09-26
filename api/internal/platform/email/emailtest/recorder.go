// Package emailtest gives other packages' tests a Sender that records messages.
package emailtest

import (
	"context"
	"sync"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/email"
)

// Recorder is an email.Sender that keeps what was sent. It validates like a real
// sender, and can be told to fail through FailWith.
type Recorder struct {
	FailWith error

	mu   sync.Mutex
	msgs []email.Message
}

func (r *Recorder) Send(_ context.Context, m email.Message) error {
	if err := m.Validate(); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.FailWith != nil {
		return r.FailWith
	}
	r.msgs = append(r.msgs, m)
	return nil
}

// Messages returns a copy of the messages sent so far.
func (r *Recorder) Messages() []email.Message {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]email.Message(nil), r.msgs...)
}

// Last returns the latest message, if any.
func (r *Recorder) Last() (email.Message, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.msgs) == 0 {
		return email.Message{}, false
	}
	return r.msgs[len(r.msgs)-1], true
}

// Reset forgets the recorded messages.
func (r *Recorder) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.msgs = nil
}
