package financeirohttp

import (
	"context"
	"mime/multipart"

	openapi_types "github.com/oapi-codegen/runtime/types"

	finapp "github.com/antoniobt12062002/torcida-jovem-platform/api/internal/financeiro/app"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/documents"
)

func comprovanteOut(d documents.Document) Comprovante {
	return Comprovante{
		Id: openapi_types.UUID(d.ID), ContentType: d.ContentType, OriginalFilename: d.OriginalFilename,
		SizeBytes: d.SizeBytes, UploadedAt: d.UploadedAt.UTC(), UploadedBy: openapi_types.UUID(d.UploadedBy), Version: d.Version,
	}
}

// multipartFile finds the "file" part of a multipart body — the only field
// CreateComprovanteMultipartBody declares — and returns its filename and
// content, ready to stream into platform/documents without buffering.
func multipartFile(body *CreateComprovanteRequestObject) (string, *multipart.Part, error) {
	for {
		part, err := body.Body.NextPart()
		if err != nil {
			return "", nil, err
		}
		if part.FormName() == "file" {
			return part.FileName(), part, nil
		}
	}
}

func (h *Handler) CreateComprovante(ctx context.Context, req CreateComprovanteRequestObject) (CreateComprovanteResponseObject, error) {
	info, err := sessionOf(ctx)
	if err != nil {
		return nil, err
	}
	filename, content, err := multipartFile(&req)
	if err != nil {
		return nil, err
	}
	doc, err := h.M.AnexarComprovante.Execute(requestContext(ctx), finapp.AnexarComprovanteInput{
		Actor: info.Principal, LancamentoID: req.Id.String(), Filename: filename, Content: content,
	})
	if err != nil {
		return nil, err
	}
	return CreateComprovante201JSONResponse(comprovanteOut(doc)), nil
}

// noStore is the fixed Cache-Control of GetComprovanteUrl's answer (a
// signed, one-time URL — never cached, CMP-02).
var noStore = "no-store"

func (h *Handler) GetComprovanteUrl(ctx context.Context, req GetComprovanteUrlRequestObject) (GetComprovanteUrlResponseObject, error) {
	info, err := sessionOf(ctx)
	if err != nil {
		return nil, err
	}
	url, err := h.M.ConsultarComprovante.Execute(requestContext(ctx), finapp.ConsultarComprovanteInput{
		Actor: info.Principal, DocumentID: req.DocumentId.String(),
	})
	if err != nil {
		return nil, err
	}
	return GetComprovanteUrl200JSONResponse{
		Body: ComprovanteUrl{Url: url}, Headers: GetComprovanteUrl200ResponseHeaders{CacheControl: &noStore},
	}, nil
}

func (h *Handler) ListComprovantes(ctx context.Context, req ListComprovantesRequestObject) (ListComprovantesResponseObject, error) {
	info, err := sessionOf(ctx)
	if err != nil {
		return nil, err
	}
	docs, err := h.M.ListarComprovantes.Execute(requestContext(ctx), finapp.ListarComprovantesInput{
		Actor: info.Principal, LancamentoID: req.Id.String(),
	})
	if err != nil {
		return nil, err
	}
	out := ComprovanteList{Items: make([]Comprovante, len(docs))}
	for i, d := range docs {
		out.Items[i] = comprovanteOut(d)
	}
	return ListComprovantes200JSONResponse(out), nil
}
