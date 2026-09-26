package testutil

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
	"github.com/getkin/kin-openapi/routers/legacy"
)

// Contract validates HTTP responses against the module contracts in
// api/openapi/. It never depends on the router under test, so it can be used
// from any package's tests.
type Contract struct {
	routers []routers.Router
}

// LoadContracts loads every module contract in api/openapi/, skipping the
// shared common.yaml, which describes components only.
func LoadContracts(t testing.TB) *Contract {
	t.Helper()
	dir := contractsDir()
	files, err := filepath.Glob(filepath.Join(dir, "*.yaml"))
	if err != nil {
		t.Fatalf("listar contratos: %v", err)
	}
	var modules []string
	for _, f := range files {
		if filepath.Base(f) != "common.yaml" {
			modules = append(modules, f)
		}
	}
	return LoadContractFiles(t, modules...)
}

// LoadContractFiles loads the given contract files, validating each document.
func LoadContractFiles(t testing.TB, paths ...string) *Contract {
	t.Helper()
	if len(paths) == 0 {
		t.Fatal("nenhum contrato para carregar")
	}
	c := &Contract{}
	for _, p := range paths {
		loader := openapi3.NewLoader()
		loader.IsExternalRefsAllowed = true
		doc, err := loader.LoadFromFile(p)
		if err != nil {
			t.Fatalf("carregar contrato %s: %v", p, err)
		}
		if err := doc.Validate(context.Background()); err != nil {
			t.Fatalf("contrato inválido %s: %v", p, err)
		}
		router, err := legacy.NewRouter(doc)
		if err != nil {
			t.Fatalf("roteador do contrato %s: %v", p, err)
		}
		c.routers = append(c.routers, router)
	}
	return c
}

// CheckResponse returns an error when the response of req does not match the
// contract: unknown route, undocumented status, content type, header or body.
func (c *Contract) CheckResponse(req *http.Request, rec *httptest.ResponseRecorder) error {
	for _, router := range c.routers {
		route, params, err := router.FindRoute(req)
		if err != nil {
			continue
		}
		in := &openapi3filter.ResponseValidationInput{
			RequestValidationInput: &openapi3filter.RequestValidationInput{
				Request:    req,
				PathParams: params,
				Route:      route,
				Options:    &openapi3filter.Options{AuthenticationFunc: openapi3filter.NoopAuthenticationFunc},
			},
			Status:  rec.Code,
			Header:  rec.Result().Header,
			Options: &openapi3filter.Options{IncludeResponseStatus: true},
		}
		in.SetBodyBytes(bytes.Clone(rec.Body.Bytes()))
		return openapi3filter.ValidateResponse(context.Background(), in)
	}
	return fmt.Errorf("nenhum contrato descreve %s %s", req.Method, req.URL.Path)
}

// ValidateResponse fails the test when the response does not match the contract.
func (c *Contract) ValidateResponse(t testing.TB, req *http.Request, rec *httptest.ResponseRecorder) {
	t.Helper()
	if err := c.CheckResponse(req, rec); err != nil {
		t.Fatalf("resposta fora do contrato (%s %s -> %d): %v", req.Method, req.URL.Path, rec.Code, err)
	}
}

func contractsDir() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "..", "..", "..", "openapi")
}
