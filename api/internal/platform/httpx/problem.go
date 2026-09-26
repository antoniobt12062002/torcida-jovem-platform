package httpx

import (
	"encoding/json"
	"errors"
	"net/http"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type FieldError struct {
	Field string `json:"field"`
	Code  string `json:"code"`
}

type problem struct {
	Type      string       `json:"type"`
	Title     string       `json:"title"`
	Status    int          `json:"status"`
	Detail    string       `json:"detail"`
	Code      string       `json:"code"`
	RequestID string       `json:"request_id"`
	Errors    []FieldError `json:"errors,omitempty"`
}

func WriteProblem(c *gin.Context, status int, code, detail string, errs ...FieldError) {
	body, err := json.Marshal(problem{
		Type:      "about:blank",
		Title:     http.StatusText(status),
		Status:    status,
		Detail:    detail,
		Code:      code,
		RequestID: RequestIDFrom(c.Request.Context()),
		Errors:    errs,
	})
	c.Abort()
	if err != nil {
		c.Status(status)
		return
	}
	c.Data(status, "application/problem+json", body)
}

// BindJSON decodes the request body into dst and validates it. On failure it
// writes the problem response and returns false.
func BindJSON(c *gin.Context, dst any) bool {
	err := c.ShouldBindJSON(dst)
	if err == nil {
		return true
	}

	var verrs validator.ValidationErrors
	var typeErr *json.UnmarshalTypeError
	switch {
	case errors.As(err, &verrs):
		fields := make([]FieldError, 0, len(verrs))
		for _, fe := range verrs {
			fields = append(fields, FieldError{Field: jsonPath(reflect.TypeOf(dst), fe.Namespace()), Code: fe.Tag()})
		}
		WriteProblem(c, http.StatusUnprocessableEntity, "validation_failed", "Um ou mais campos são inválidos.", fields...)
	case errors.As(err, &typeErr):
		field := typeErr.Field
		if field == "" {
			field = "body"
		}
		WriteProblem(c, http.StatusUnprocessableEntity, "validation_failed", "Um ou mais campos são inválidos.",
			FieldError{Field: field, Code: "invalid_type"})
	default:
		WriteProblem(c, http.StatusBadRequest, "invalid_json", "O corpo da requisição não é um JSON válido.")
	}
	return false
}

// jsonPath turns a validator namespace such as "signup.Address.Zip" into the
// JSON path clients see, such as "address.zip".
func jsonPath(t reflect.Type, namespace string) string {
	parts := strings.Split(namespace, ".")
	if len(parts) > 1 {
		parts = parts[1:]
	}
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		name, index := part, ""
		if i := strings.Index(part, "["); i >= 0 {
			name, index = part[:i], part[i:]
		}
		t = indirect(t)
		if t == nil || t.Kind() != reflect.Struct {
			out = append(out, part)
			continue
		}
		f, ok := t.FieldByName(name)
		if !ok {
			out = append(out, part)
			t = nil
			continue
		}
		jsonName := strings.Split(f.Tag.Get("json"), ",")[0]
		if jsonName == "" || jsonName == "-" {
			jsonName = name
		}
		out = append(out, jsonName+index)
		t = f.Type
	}
	return strings.Join(out, ".")
}

func indirect(t reflect.Type) reflect.Type {
	for t != nil && (t.Kind() == reflect.Pointer || t.Kind() == reflect.Slice || t.Kind() == reflect.Array) {
		t = t.Elem()
	}
	return t
}
