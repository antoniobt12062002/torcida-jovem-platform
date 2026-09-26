package emailtest

import (
	"context"
	"sync"
	"testing"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/email"
)

var _ email.Sender = (*Recorder)(nil)

func msg(to string) email.Message {
	return email.Message{To: to, Subject: "Assunto", TextBody: "Corpo"}
}

func TestRecorderKeepsTheSentMessagesInOrder(t *testing.T) {
	r := &Recorder{}

	_ = r.Send(context.Background(), msg("a@x.com"))
	_ = r.Send(context.Background(), msg("b@x.com"))

	got := r.Messages()
	if len(got) != 2 || got[0].To != "a@x.com" || got[1].To != "b@x.com" {
		t.Errorf("mensagens = %+v", got)
	}
	if last, ok := r.Last(); !ok || last.To != "b@x.com" {
		t.Errorf("Last = %+v, %v", last, ok)
	}
}

func TestRecorderValidatesLikeARealSender(t *testing.T) {
	r := &Recorder{}

	if err := r.Send(context.Background(), email.Message{To: "inválido", Subject: "s", TextBody: "b"}); err == nil {
		t.Error("uma mensagem inválida deveria ser recusada")
	}
	if n := len(r.Messages()); n != 0 {
		t.Errorf("nada deveria ter sido guardado, n = %d", n)
	}
	if _, ok := r.Last(); ok {
		t.Error("Last sem mensagens deveria devolver ok=false")
	}
}

func TestRecorderCanFailOnDemandAndReset(t *testing.T) {
	r := &Recorder{FailWith: context.DeadlineExceeded}

	if err := r.Send(context.Background(), msg("a@x.com")); err != context.DeadlineExceeded {
		t.Errorf("err = %v", err)
	}
	if len(r.Messages()) != 0 {
		t.Error("uma falha simulada não guarda a mensagem")
	}
	r.FailWith = nil
	_ = r.Send(context.Background(), msg("a@x.com"))
	r.Reset()
	if len(r.Messages()) != 0 {
		t.Error("Reset deveria limpar")
	}
}

func TestRecorderIsSafeForConcurrentUse(t *testing.T) {
	r := &Recorder{}
	var wg sync.WaitGroup
	for range 50 {
		wg.Go(func() { _ = r.Send(context.Background(), msg("a@x.com")) })
	}
	wg.Wait()

	if n := len(r.Messages()); n != 50 {
		t.Errorf("n = %d", n)
	}
}
