package artworkcontroller

import (
	"context"
	"net/http"

	"github.com/tombell/memoir/internal/api/payload"
	"github.com/tombell/memoir/internal/controllers"
	"github.com/tombell/memoir/internal/errors"
	"github.com/tombell/memoir/internal/stores/artworkstore"
)

// CreateRequest defines the data to read from the HTTP request.
type CreateRequest struct {
	Artwork *payload.File `file:"artwork"`
}

// CreateResponse defines the data to write to the HTTP response.
type CreateResponse struct {
	Upload *artworkstore.Upload `json:"data"`

	status int
}

// StatusCode returns the status code to use for the HTTP response.
func (r *CreateResponse) StatusCode() int {
	return r.status
}

// Create returns an action function for uploading artwork using the artwork
// store.
func Create(artworkStore *artworkstore.Store) controllers.ActionFunc[CreateRequest, *CreateResponse] {
	return func(ctx context.Context, input CreateRequest) (*CreateResponse, error) {
		if input.Artwork == nil {
			return nil, errors.E("artwork[create]", http.StatusBadRequest, errors.M{"artwork": {"an image file is required"}})
		}
		defer input.Artwork.File.Close()
		if input.Artwork.Header.Size > 8<<20 {
			return nil, errors.E("artwork[create]", http.StatusRequestEntityTooLarge)
		}
		upload, exists, err := artworkStore.Upload(ctx, input.Artwork.File)
		if err != nil {
			return nil, err
		}

		resp := &CreateResponse{
			Upload: upload,
			status: http.StatusCreated,
		}

		if exists {
			resp.status = http.StatusOK
		}

		return resp, nil
	}
}
