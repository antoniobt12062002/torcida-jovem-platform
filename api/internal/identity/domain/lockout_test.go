package domain

import (
	"testing"
	"time"
)

func failuresAt(base time.Time, offsets ...time.Duration) []time.Time {
	out := make([]time.Time, len(offsets))
	for i, o := range offsets {
		out[i] = base.Add(o)
	}
	return out
}

func TestLockoutLimitsAreFiveFailuresInFifteenMinutesLockedForFifteen(t *testing.T) {
	if MaxLoginFailures != 5 || LoginFailureWindow != 15*time.Minute || LoginLockDuration != 15*time.Minute {
		t.Errorf("limites = %d / %v / %v", MaxLoginFailures, LoginFailureWindow, LoginLockDuration)
	}
}

// IDN-02.8: a quinta falha seguida em 15 minutos bloqueia por 15 minutos.
func TestLockoutBlocksAtTheFifthFailureWithinTheWindow(t *testing.T) {
	f := failuresAt(now, 0, time.Minute, 2*time.Minute, 3*time.Minute)

	if _, blocked := Lockout(f, now.Add(3*time.Minute)); blocked {
		t.Error("4 falhas não bloqueiam")
	}
	f = append(f, now.Add(4*time.Minute))
	until, blocked := Lockout(f, now.Add(4*time.Minute))

	if !blocked || !until.Equal(now.Add(19*time.Minute)) {
		t.Errorf("a quinta falha bloqueia até 15 minutos depois dela: blocked = %v, until = %v", blocked, until)
	}
}

func TestLockoutEndsFifteenMinutesAfterTheFifthFailure(t *testing.T) {
	f := failuresAt(now, 0, time.Minute, 2*time.Minute, 3*time.Minute, 4*time.Minute)

	if _, blocked := Lockout(f, now.Add(19*time.Minute-time.Second)); !blocked {
		t.Error("ainda bloqueado 1 segundo antes do fim")
	}
	if _, blocked := Lockout(f, now.Add(19*time.Minute)); blocked {
		t.Error("o bloqueio termina exatamente 15 minutos depois da quinta falha")
	}
	if _, blocked := Lockout(f, now.Add(3*time.Hour)); blocked {
		t.Error("nunca há bloqueio permanente")
	}
}

func TestLockoutIgnoresFailuresSpreadOverMoreThanTheWindow(t *testing.T) {
	// 5 falhas, mas a primeira e a quinta distam mais de 15 minutos.
	f := failuresAt(now, 0, 4*time.Minute, 8*time.Minute, 12*time.Minute, 16*time.Minute)

	if _, blocked := Lockout(f, now.Add(16*time.Minute)); blocked {
		t.Error("5 falhas espalhadas em mais de 15 minutos não bloqueiam")
	}
	edge := failuresAt(now, 0, 4*time.Minute, 8*time.Minute, 12*time.Minute, 15*time.Minute)
	if _, blocked := Lockout(edge, now.Add(15*time.Minute)); !blocked {
		t.Error("5 falhas em exatamente 15 minutos bloqueiam")
	}
}

func TestLockoutLooksAtTheLastFiveFailures(t *testing.T) {
	// Um bloqueio antigo, seguido de mais falhas: só as cinco últimas contam.
	f := failuresAt(now, 0, time.Second, 2*time.Second, 3*time.Second, 4*time.Second, 30*time.Minute)

	if _, blocked := Lockout(f, now.Add(30*time.Minute)); blocked {
		t.Error("as cinco últimas (uma delas 30 minutos depois) não estão na mesma janela")
	}
	if _, blocked := Lockout(nil, now); blocked {
		t.Error("sem falhas não há bloqueio")
	}
}

// Uma sequência nova de falhas bloqueia de novo, depois de um bloqueio antigo já expirado.
func TestLockoutBlocksAgainForTheLatestBurstOfFailures(t *testing.T) {
	f := failuresAt(now, 0, time.Minute, 2*time.Minute, 3*time.Minute, 4*time.Minute,
		20*time.Minute, 21*time.Minute, 22*time.Minute, 23*time.Minute, 24*time.Minute)

	until, blocked := Lockout(f, now.Add(25*time.Minute))

	if !blocked || !until.Equal(now.Add(39*time.Minute)) {
		t.Errorf("blocked = %v, until = %v", blocked, until)
	}
}
