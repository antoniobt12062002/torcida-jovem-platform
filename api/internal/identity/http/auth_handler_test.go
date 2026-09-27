//go:build integration

package identityhttp_test

import (
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/domain"
)

func sessionCookieOf(w *httptest.ResponseRecorder) *http.Cookie {
	for _, c := range w.Result().Cookies() {
		if c.Name == "tj_session" {
			return c
		}
	}
	return nil
}

func loginBody(u domain.User, pw string) map[string]string {
	return map[string]string{"email": u.Email, "password": pw}
}

func keys(m map[string]any) []string {
	var out []string
	for k := range m {
		out = append(out, k)
	}
	slices.Sort(out)
	return out
}

// IDN-02.1 e .9: o login devolve o mesmo corpo do `me` e define o cookie com os atributos de segurança.
func TestLoginSetsTheSessionCookieAndReturnsTheSameContextAsMe(t *testing.T) {
	e := newEnv(t)
	u := e.newUser(domain.RolePresidente)

	w := e.do(http.MethodPost, "/api/v1/auth/login", loginBody(u, goodPassword))

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, corpo = %s", w.Code, w.Body.String())
	}
	c := sessionCookieOf(w)
	if c == nil || c.Value == "" || !c.HttpOnly || !c.Secure || c.SameSite != http.SameSiteLaxMode || c.Path != "/" || c.MaxAge != 8*3600 {
		t.Fatalf("cookie = %+v", c)
	}
	body := decode(t, w)
	me := e.do(http.MethodGet, "/api/v1/auth/me", nil, cookie(c.Value))
	if me.Code != http.StatusOK || strings.TrimSpace(me.Body.String()) != strings.TrimSpace(w.Body.String()) {
		t.Errorf("o login e o me deveriam devolver o mesmo corpo:\nlogin: %s\nme:    %s", w.Body.String(), me.Body.String())
	}
	if want := []string{"admin_membership", "csrf_token", "must_change_password", "permissions", "roles", "user"}; !slices.Equal(keys(body), want) {
		t.Errorf("campos = %v", keys(body))
	}
	user := body["user"].(map[string]any)
	if user["email"] != u.Email || user["id"] != u.ID || user["name"] != u.Name || len(user) != 3 {
		t.Errorf("user = %v", user)
	}
	if roles := body["roles"].([]any); len(roles) != 2 || roles[0] != "ASSOCIADO" || roles[1] != "PRESIDENTE" {
		t.Errorf("roles = %v", roles)
	}
	if len(body["permissions"].([]any)) == 0 || body["csrf_token"] == "" || body["must_change_password"] != false {
		t.Errorf("permissões, csrf ou flag: %v", body)
	}
	for _, leak := range []string{"hash", "argon2", "secret"} {
		if strings.Contains(strings.ToLower(w.Body.String()), leak) {
			t.Errorf("a resposta do login não pode conter %q", leak)
		}
	}
}

// IDN-02.9: o vínculo administrativo aparece só com motivo e data, ou nulo.
func TestMeShowsTheMembershipSummaryOnlyWithReasonAndDateOrNull(t *testing.T) {
	e := newEnv(t)
	_, admin := e.signedIn(domain.RoleTesouraria)
	_, plain := e.signedIn()

	withMembership := decode(t, e.as(admin, http.MethodGet, "/api/v1/auth/me", nil))
	without := decode(t, e.as(plain, http.MethodGet, "/api/v1/auth/me", nil))

	m := withMembership["admin_membership"].(map[string]any)
	if !slices.Equal(keys(m), []string{"granted_at", "reason"}) || m["reason"] != "membro da diretoria eleita" {
		t.Errorf("admin_membership = %v", m)
	}
	if v, present := without["admin_membership"]; !present || v != nil {
		t.Errorf("sem vínculo, admin_membership deveria ser null: %v (presente: %v)", v, present)
	}
}

// IDN-02.3 e .4: e-mail inexistente, senha errada e usuário inativo devolvem a mesma resposta.
func TestLoginErrorsAreIdenticalForUnknownEmailWrongPasswordAndInactiveUser(t *testing.T) {
	e := newEnv(t)
	known := e.newUser()
	inactive := e.newUser()
	if err := e.mod.Users.SetActive(t.Context(), inactive.ID, false); err != nil {
		t.Fatal(err)
	}

	cases := map[string]map[string]string{
		"e-mail inexistente": {"email": "ninguem@exemplo.com", "password": goodPassword},
		"senha errada":       loginBody(known, "senha-errada-123"),
		"usuário inativo":    loginBody(inactive, goodPassword),
	}
	var first map[string]any
	for name, body := range cases {
		w := e.do(http.MethodPost, "/api/v1/auth/login", body)
		if w.Code != http.StatusUnauthorized || codeOf(t, w) != "invalid_credentials" {
			t.Fatalf("%s: status = %d, corpo = %s", name, w.Code, w.Body.String())
		}
		got := decode(t, w)
		delete(got, "request_id")
		if first == nil {
			first = got
		} else if !equalMaps(first, got) {
			t.Errorf("%s: corpo diferente dos outros: %v x %v", name, got, first)
		}
		if sessionCookieOf(w) != nil {
			t.Errorf("%s: um login recusado não define cookie", name)
		}
	}
}

func equalMaps(a, b map[string]any) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if b[k] != v {
			return false
		}
	}
	return true
}

// IDN-02.8: cinco falhas seguidas bloqueiam o e-mail, e o 429 diz quando tentar de novo.
func TestLoginIsBlockedAfterFiveFailuresAndSaysWhenToRetry(t *testing.T) {
	e := newEnv(t)
	u := e.newUser()

	for i := 0; i < 5; i++ {
		if w := e.do(http.MethodPost, "/api/v1/auth/login", loginBody(u, "senha-errada-123")); w.Code != http.StatusUnauthorized {
			t.Fatalf("falha %d: status = %d", i+1, w.Code)
		}
	}
	blocked := e.do(http.MethodPost, "/api/v1/auth/login", loginBody(u, goodPassword))

	if blocked.Code != http.StatusTooManyRequests || codeOf(t, blocked) != "login_blocked" {
		t.Fatalf("status = %d, corpo = %s", blocked.Code, blocked.Body.String())
	}
	if ra := blocked.Header().Get("Retry-After"); ra == "" || ra == "0" {
		t.Errorf("Retry-After = %q", ra)
	}
	e.now = e.now.Add(16 * time.Minute)
	if w := e.do(http.MethodPost, "/api/v1/auth/login", loginBody(u, goodPassword)); w.Code != http.StatusOK {
		t.Errorf("passado o bloqueio, o login volta a funcionar: %d", w.Code)
	}
}

// IDN-02.6: o logout revoga a sessão e limpa o cookie.
func TestLogoutRevokesTheSessionAndClearsTheCookie(t *testing.T) {
	e := newEnv(t)
	_, s := e.signedIn()

	w := e.as(s, http.MethodPost, "/api/v1/auth/logout", nil)

	if w.Code != http.StatusNoContent {
		t.Fatalf("status = %d, corpo = %s", w.Code, w.Body.String())
	}
	if c := sessionCookieOf(w); c == nil || c.Value != "" || c.MaxAge >= 0 || !c.HttpOnly {
		t.Errorf("o cookie deveria ser limpo: %+v", c)
	}
	if after := e.as(s, http.MethodGet, "/api/v1/auth/me", nil); after.Code != http.StatusUnauthorized || codeOf(t, after) != "unauthenticated" {
		t.Errorf("a sessão revogada não vale mais: %d %s", after.Code, after.Body.String())
	}
	if len(e.events("auth.logout")) != 1 {
		t.Error("o logout deveria gerar auth.logout")
	}
}

// IDN-03: rotas autenticadas sem sessão respondem 401 em problem+json.
func TestAuthenticatedRoutesWithoutASessionAnswer401(t *testing.T) {
	e := newEnv(t)

	for _, r := range []struct{ method, path string }{{http.MethodGet, "/api/v1/auth/me"}, {http.MethodPost, "/api/v1/auth/logout"}, {http.MethodPost, "/api/v1/auth/password"}} {
		w := e.do(r.method, r.path, nil)
		if w.Code != http.StatusUnauthorized || codeOf(t, w) != "unauthenticated" || decode(t, w)["request_id"] == "" {
			t.Errorf("%s %s: %d %s", r.method, r.path, w.Code, w.Body.String())
		}
	}
}

// IDN-04.5: a troca de senha mantém a sessão atual, revoga as outras e libera o acesso.
func TestChangePasswordKeepsTheCurrentSessionAndRevokesTheOthers(t *testing.T) {
	e := newEnv(t)
	u, current := e.signedIn()
	other := e.login(u)

	w := e.as(current, http.MethodPost, "/api/v1/auth/password", map[string]string{
		"current_password": goodPassword, "new_password": "outra-senha-forte-456",
	})

	if w.Code != http.StatusNoContent {
		t.Fatalf("status = %d, corpo = %s", w.Code, w.Body.String())
	}
	if e.as(current, http.MethodGet, "/api/v1/auth/me", nil).Code != http.StatusOK {
		t.Error("a sessão que trocou a senha deveria continuar")
	}
	if e.as(other, http.MethodGet, "/api/v1/auth/me", nil).Code != http.StatusUnauthorized {
		t.Error("as outras sessões deveriam ser revogadas")
	}
	if e.do(http.MethodPost, "/api/v1/auth/login", loginBody(u, "outra-senha-forte-456")).Code != http.StatusOK {
		t.Error("a senha nova deveria entrar")
	}
	if len(e.events("user.password_change")) != 1 {
		t.Error("a troca deveria gerar user.password_change")
	}
}

// IDN-04.6: a senha atual errada é 403 invalid_current_password.
func TestChangePasswordWithTheWrongCurrentPasswordIs403(t *testing.T) {
	e := newEnv(t)
	_, s := e.signedIn()

	w := e.as(s, http.MethodPost, "/api/v1/auth/password", map[string]string{
		"current_password": "senha-errada-123", "new_password": "outra-senha-forte-456",
	})

	if w.Code != http.StatusForbidden || codeOf(t, w) != "invalid_current_password" {
		t.Errorf("status = %d, corpo = %s", w.Code, w.Body.String())
	}
}

// IDN-07.1: a solicitação responde 202 com o mesmo corpo exista a conta ou não, e só envia e-mail se existir.
func TestRecoveryRequestAnswersTheSameWhetherOrNotTheAccountExists(t *testing.T) {
	e := newEnv(t)
	u := e.newUser()

	exists := e.do(http.MethodPost, "/api/v1/auth/password-reset/request", map[string]string{"email": u.Email})
	sent := len(e.mail.Messages())
	unknown := e.do(http.MethodPost, "/api/v1/auth/password-reset/request", map[string]string{"email": "ninguem@exemplo.com"})

	if exists.Code != http.StatusAccepted || unknown.Code != http.StatusAccepted {
		t.Fatalf("status = %d e %d", exists.Code, unknown.Code)
	}
	if exists.Body.String() != unknown.Body.String() {
		t.Errorf("os corpos deveriam ser idênticos:\n%s\n%s", exists.Body.String(), unknown.Body.String())
	}
	if msg := decode(t, exists)["message"]; msg != "Se existir uma conta vinculada ao e-mail informado, enviaremos instruções." {
		t.Errorf("message = %v", msg)
	}
	if sent != 1 || len(e.mail.Messages()) != 1 {
		t.Errorf("só a conta existente recebe e-mail: %d antes, %d depois", sent, len(e.mail.Messages()))
	}
}

// IDN-07.5: o token redefine a senha, revoga as sessões antigas e vale uma vez.
func TestRecoveryConfirmSetsTheNewPasswordRevokesSessionsAndTokenWorksOnce(t *testing.T) {
	e := newEnv(t)
	u, old := e.signedIn()
	e.do(http.MethodPost, "/api/v1/auth/password-reset/request", map[string]string{"email": u.Email})
	token := e.tokenFromMail()

	w := e.do(http.MethodPost, "/api/v1/auth/password-reset/confirm", map[string]string{"token": token, "new_password": "senha-nova-recuperada-1"})

	if w.Code != http.StatusNoContent {
		t.Fatalf("status = %d, corpo = %s", w.Code, w.Body.String())
	}
	if e.as(old, http.MethodGet, "/api/v1/auth/me", nil).Code != http.StatusUnauthorized {
		t.Error("as sessões antigas deveriam cair")
	}
	if e.do(http.MethodPost, "/api/v1/auth/login", loginBody(u, "senha-nova-recuperada-1")).Code != http.StatusOK {
		t.Error("a senha nova deveria entrar")
	}
	again := e.do(http.MethodPost, "/api/v1/auth/password-reset/confirm", map[string]string{"token": token, "new_password": "outra-senha-recuperada-2"})
	if again.Code != http.StatusBadRequest || codeOf(t, again) != "invalid_reset_token" {
		t.Errorf("o token usado não vale: %d %s", again.Code, again.Body.String())
	}
}

// IDN-07.6: token desconhecido, expirado e usado dão a mesma resposta 400.
func TestRecoveryConfirmAnswersIdenticallyForUnknownExpiredAndUsedTokens(t *testing.T) {
	e := newEnv(t)
	u := e.newUser()
	e.do(http.MethodPost, "/api/v1/auth/password-reset/request", map[string]string{"email": u.Email})
	expired := e.tokenFromMail()
	e.now = e.now.Add(31 * time.Minute)
	e.do(http.MethodPost, "/api/v1/auth/password-reset/request", map[string]string{"email": u.Email})
	used := e.tokenFromMail()
	e.do(http.MethodPost, "/api/v1/auth/password-reset/confirm", map[string]string{"token": used, "new_password": "senha-nova-recuperada-1"})

	var first map[string]any
	for name, token := range map[string]string{"desconhecido": strings.Repeat("a", 43), "expirado": expired, "usado": used} {
		w := e.do(http.MethodPost, "/api/v1/auth/password-reset/confirm", map[string]string{"token": token, "new_password": "senha-nova-recuperada-2"})
		if w.Code != http.StatusBadRequest || codeOf(t, w) != "invalid_reset_token" {
			t.Fatalf("%s: %d %s", name, w.Code, w.Body.String())
		}
		got := decode(t, w)
		delete(got, "request_id")
		if first == nil {
			first = got
		} else if !equalMaps(first, got) {
			t.Errorf("%s: corpo diferente: %v x %v", name, got, first)
		}
	}
	if len(e.events("auth.password_reset_failed")) != 3 {
		t.Errorf("cada tentativa recusada deveria gerar auth.password_reset_failed: %d", len(e.events("auth.password_reset_failed")))
	}
}

// IDN-07.7: senha fora da política dá 422 com o código da política e o token continua valendo.
func TestRecoveryConfirmWithAWeakPasswordIs422AndKeepsTheToken(t *testing.T) {
	e := newEnv(t)
	u := e.newUser()
	e.do(http.MethodPost, "/api/v1/auth/password-reset/request", map[string]string{"email": u.Email})
	token := e.tokenFromMail()

	weak := e.do(http.MethodPost, "/api/v1/auth/password-reset/confirm", map[string]string{"token": token, "new_password": "curta"})

	if weak.Code != http.StatusUnprocessableEntity || codeOf(t, weak) != "password_too_short" {
		t.Fatalf("status = %d, corpo = %s", weak.Code, weak.Body.String())
	}
	if ok := e.do(http.MethodPost, "/api/v1/auth/password-reset/confirm", map[string]string{"token": token, "new_password": "senha-nova-recuperada-1"}); ok.Code != http.StatusNoContent {
		t.Errorf("o token deveria continuar válido: %d %s", ok.Code, ok.Body.String())
	}
}

// IDN-07.10: as rotas públicas aplicam a checagem de origem; sem sessão, não exigem CSRF.
func TestPublicRoutesCheckTheOriginButNeitherTheSessionNorCSRF(t *testing.T) {
	e := newEnv(t)
	u := e.newUser()

	for _, route := range []struct {
		path string
		body any
	}{
		{"/api/v1/auth/login", loginBody(u, goodPassword)},
		{"/api/v1/auth/password-reset/request", map[string]string{"email": u.Email}},
		{"/api/v1/auth/password-reset/confirm", map[string]string{"token": "x", "new_password": "senha-nova-recuperada-1"}},
	} {
		blocked := e.do(http.MethodPost, route.path, route.body, origin("https://evil.example"))
		if blocked.Code != http.StatusForbidden || codeOf(t, blocked) != "origin_not_allowed" {
			t.Errorf("%s com origem estranha: %d %s", route.path, blocked.Code, blocked.Body.String())
		}
		if allowed := e.do(http.MethodPost, route.path, route.body); allowed.Code == http.StatusForbidden || allowed.Code == http.StatusUnauthorized && route.path != "/api/v1/auth/login" {
			t.Errorf("%s com a origem permitida não deveria pedir sessão nem CSRF: %d", route.path, allowed.Code)
		}
	}
}

// IDN-03.1 e IDN-02: a troca de senha obrigatória libera só logout, me e a troca.
func TestMustChangePasswordAllowsOnlyLogoutMeAndTheChange(t *testing.T) {
	e := newEnv(t)
	u := e.newUser()
	if err := e.mod.Users.SetMustChangePassword(t.Context(), u.ID, true); err != nil {
		t.Fatal(err)
	}
	s := e.login(u)

	if w := e.as(s, http.MethodGet, "/api/v1/auth/me", nil); w.Code != http.StatusOK || decode(t, w)["must_change_password"] != true {
		t.Errorf("me: %d %s", w.Code, w.Body.String())
	}
	blocked := e.as(s, http.MethodGet, "/api/v1/users", nil)
	if blocked.Code != http.StatusForbidden || codeOf(t, blocked) != "password_change_required" {
		t.Errorf("listar usuários: %d %s", blocked.Code, blocked.Body.String())
	}
	change := e.as(s, http.MethodPost, "/api/v1/auth/password", map[string]string{"current_password": goodPassword, "new_password": "outra-senha-forte-456"})
	if change.Code != http.StatusNoContent {
		t.Fatalf("troca: %d %s", change.Code, change.Body.String())
	}
	if w := e.as(s, http.MethodGet, "/api/v1/auth/me", nil); decode(t, w)["must_change_password"] != false {
		t.Error("depois da troca a flag deveria cair")
	}
}

// IDN-04.12: cinco senhas atuais erradas bloqueiam a troca, com 429 e Retry-After.
func TestChangePasswordIsBlockedAfterFiveWrongCurrentPasswords(t *testing.T) {
	e := newEnv(t)
	_, s := e.signedIn()
	wrong := map[string]string{"current_password": "senha-errada-123", "new_password": "outra-senha-forte-456"}

	for i := 0; i < 5; i++ {
		if w := e.as(s, http.MethodPost, "/api/v1/auth/password", wrong); w.Code != http.StatusForbidden {
			t.Fatalf("tentativa %d: status = %d", i+1, w.Code)
		}
	}
	blocked := e.as(s, http.MethodPost, "/api/v1/auth/password", map[string]string{"current_password": goodPassword, "new_password": "outra-senha-forte-456"})

	if blocked.Code != http.StatusTooManyRequests || codeOf(t, blocked) != "password_change_blocked" {
		t.Fatalf("status = %d, corpo = %s", blocked.Code, blocked.Body.String())
	}
	if ra := blocked.Header().Get("Retry-After"); ra == "" || ra == "0" {
		t.Errorf("Retry-After = %q", ra)
	}
}
