package httpx

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
	"github.com/getkin/kin-openapi/routers/legacy"
	"github.com/gin-gonic/gin"
)

// kin-openapi only checks the formats it knows; register the ones the
// contracts use so a format violation is a 422 and not silently accepted.
func init() {
	openapi3.DefineStringFormatValidator("uuid", openapi3.NewRegexpFormatValidator(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`))
	openapi3.DefineStringFormatValidator("email", openapi3.NewRegexpFormatValidator(`^[^@\s]+@[^@\s]+\.[^@\s]+$`))
}

// LoadContract reads and validates an OpenAPI document, resolving references
// to the shared common.yaml.
func LoadContract(path string) (*openapi3.T, error) {
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	doc, err := loader.LoadFromFile(path)
	if err != nil {
		return nil, err
	}
	if err := doc.Validate(context.Background()); err != nil {
		return nil, err
	}
	return doc, nil
}

// NewContractValidator validates each request against the module contracts:
// structure, types, formats, enums and required fields of the body, query and
// path. It carries no business rule; those stay in the domain modules.
//
// A request that matches no operation is passed on untouched. Rejecting such
// routes in production is a future decision (see .specs/STATE.md).
func NewContractValidator(docs ...*openapi3.T) (gin.HandlerFunc, error) {
	if len(docs) == 0 {
		return nil, errors.New("httpx: nenhum contrato para validar")
	}
	routerList := make([]routers.Router, 0, len(docs))
	for _, doc := range docs {
		r, err := legacy.NewRouter(doc)
		if err != nil {
			return nil, err
		}
		routerList = append(routerList, r)
	}
	options := &openapi3filter.Options{
		MultiError:         true,
		AuthenticationFunc: openapi3filter.NoopAuthenticationFunc,
	}
	return func(c *gin.Context) {
		for _, router := range routerList {
			route, params, err := router.FindRoute(c.Request)
			if err != nil {
				continue
			}
			err = openapi3filter.ValidateRequest(c.Request.Context(), &openapi3filter.RequestValidationInput{
				Request: c.Request, PathParams: params, Route: route, Options: options,
			})
			if err != nil {
				writeContractProblem(c, err)
				return
			}
			break
		}
		c.Next()
	}, nil
}

func writeContractProblem(c *gin.Context, err error) {
	fields, malformed := fieldErrors(err, nil)
	if malformed {
		WriteProblem(c, http.StatusBadRequest, "invalid_json", "O corpo da requisição não é um JSON válido.")
		return
	}
	code, detail := "validation_failed", "Um ou mais campos são inválidos."
	if limitOutOfRange(fields) {
		code, detail = "invalid_limit", "O limite deve estar entre 1 e 100."
	}
	WriteProblem(c, http.StatusUnprocessableEntity, code, detail, fields...)
}

// limitOutOfRange reports a `limit` (API-02.5) below its minimum or above its maximum.
func limitOutOfRange(fields []FieldError) bool {
	for _, f := range fields {
		if f.Field == "limit" && (f.Code == "too_small" || f.Code == "too_large") {
			return true
		}
	}
	return false
}

// fieldErrors flattens a kin-openapi error into per-field errors. malformed is
// true when the body is not parseable JSON. from is the enclosing request error,
// which says whether the failure is in a parameter or in the body.
func fieldErrors(err error, from *openapi3filter.RequestError) (fields []FieldError, malformed bool) {
	switch e := err.(type) {
	case openapi3.MultiError:
		for _, inner := range e {
			f, m := fieldErrors(inner, from)
			fields = append(fields, f...)
			malformed = malformed || m
		}
		return fields, malformed
	case *openapi3filter.RequestError:
		return fieldErrors(e.Err, e)
	case *openapi3.SchemaError:
		return []FieldError{schemaFieldError(from, e)}, false
	case *openapi3filter.ParseError:
		if from != nil && from.RequestBody != nil {
			return nil, true
		}
		return []FieldError{{Field: paramName(from), Code: "invalid_type"}}, false
	}
	if from != nil && from.RequestBody != nil && from.Parameter == nil && err != nil {
		if _, ok := err.(interface{ Unwrap() error }); !ok || strings.Contains(err.Error(), "value is required") {
			return []FieldError{{Field: "body", Code: "required"}}, false
		}
		return nil, true
	}
	return []FieldError{{Field: paramName(from), Code: "invalid_value"}}, false
}

func paramName(e *openapi3filter.RequestError) string {
	if e != nil && e.Parameter != nil {
		return e.Parameter.Name
	}
	return "body"
}

func schemaFieldError(reqErr *openapi3filter.RequestError, e *openapi3.SchemaError) FieldError {
	field := strings.Join(e.JSONPointer(), ".")
	if reqErr != nil && reqErr.Parameter != nil {
		field = reqErr.Parameter.Name
	}
	if field == "" {
		field = "body"
	}
	return FieldError{Field: field, Code: schemaCode(e.SchemaField)}
}

func schemaCode(schemaField string) string {
	switch schemaField {
	case "required":
		return "required"
	case "type":
		return "invalid_type"
	case "enum":
		return "invalid_enum"
	case "format", "pattern":
		return "invalid_format"
	case "minLength", "minItems", "minimum", "exclusiveMinimum":
		return "too_small"
	case "maxLength", "maxItems", "maximum", "exclusiveMaximum":
		return "too_large"
	}
	return "invalid_value"
}
