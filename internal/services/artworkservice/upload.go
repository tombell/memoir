package artworkservice

import (
	"context"
	"net/http"

	"github.com/tombell/memoir/internal/api/payload"
	"github.com/tombell/memoir/internal/services"
	"github.com/tombell/memoir/internal/stores/artworkstore"
)

type UploadRequest struct {
	Artwork *payload.File `file:"artwork"`
}

type UploadResponse struct {
	Upload *artworkstore.Upload `json:"upload"`

	status int
}

func (r *UploadResponse) StatusCode() int {
	return r.status
}

func Upload(artworkStore *artworkstore.Store) services.ServiceFunc[UploadRequest, *UploadResponse] {
	return func(ctx context.Context, input UploadRequest) (*UploadResponse, error) {
		upload, exists, err := artworkStore.Upload(ctx, input.Artwork.File, input.Artwork.Header.Filename)
		if err != nil {
			return nil, err
		}

		resp := &UploadResponse{
			Upload: upload,
			status: http.StatusCreated,
		}

		if exists {
			resp.status = http.StatusOK
		}

		return resp, nil
	}
}
