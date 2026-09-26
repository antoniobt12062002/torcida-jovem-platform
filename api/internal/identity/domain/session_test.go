package domain

import (
	"crypto/sha256"
	"encoding/base64"
	"testing"
	"time"
)

func TestNewSessionTokenIs256BitsRandomBase64URLWithItsSHA256(t *testing.T) {
	token, hash, err := NewSessionToken()
	if err != nil {
		t.Fatal(err)
	}

	raw, decodeErr := base64.RawURLEncoding.DecodeString(token)
	if decodeErr != nil || len(raw) != 32 {
		t.Errorf("o token deveria ter 256 bits em base64url: len = %d, err = %v", len(raw), decodeErr)
	}
	want := sha256.Sum256([]byte(token))
	if string(hash) != string(want[:]) || string(HashSessionToken(token)) != string(want[:]) {
		t.Error("o hash deveria ser o SHA-256 do token")
	}
}

func TestSessionTokensAreUnique(t *testing.T) {
	seen := map[string]bool{}
	for range 200 {
		token, _, _ := NewSessionToken()
		if seen[token] {
			t.Fatal("token repetido")
		}
		seen[token] = true
	}
}

func TestNewCSRFTokenIs256BitsAndDifferentEachTime(t *testing.T) {
	a, err := NewCSRFToken()
	if err != nil {
		t.Fatal(err)
	}
	b, _ := NewCSRFToken()

	raw, _ := base64.RawURLEncoding.DecodeString(a)
	if len(raw) != 32 || a == b {
		t.Errorf("len = %d, iguais = %v", len(raw), a == b)
	}
}

const (
	idle     = 60 * time.Minute
	absolute = 8 * time.Hour
)

// IDN-02.5: mais de 60 minutos sem uso, ou mais de 8 horas de idade, expira.
func TestSessionExpiresAfterTheIdleLimit(t *testing.T) {
	s := Session{CreatedAt: now, LastSeenAt: now, ExpiresAt: now.Add(absolute)}

	if s.Expired(now.Add(idle), idle) {
		t.Error("exatamente 60 minutos sem uso ainda vale (o limite é mais de 60)")
	}
	if !s.Expired(now.Add(idle+time.Second), idle) {
		t.Error("mais de 60 minutos sem uso não vale mais")
	}
}

func TestSessionExpiresAfterTheAbsoluteLimitEvenWhenRecentlyUsed(t *testing.T) {
	s := Session{CreatedAt: now, LastSeenAt: now.Add(absolute - time.Minute), ExpiresAt: now.Add(absolute)}

	if s.Expired(now.Add(absolute), idle) {
		t.Error("exatamente 8 horas ainda vale")
	}
	if !s.Expired(now.Add(absolute+time.Second), idle) {
		t.Error("depois do teto absoluto não vale, mesmo com uso recente")
	}
}

func TestRevokedSessionIsNeverValid(t *testing.T) {
	revoked := now
	s := Session{CreatedAt: now, LastSeenAt: now, ExpiresAt: now.Add(absolute), RevokedAt: &revoked}

	if !s.Expired(now, idle) {
		t.Error("sessão revogada não vale")
	}
}
