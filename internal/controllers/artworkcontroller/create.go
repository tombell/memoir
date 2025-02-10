package artworkcontroller

import (
	"context"
	"net/http"

	"github.com/tombell/memoir/internal/api/payload"
	"github.com/tombell/memoir/internal/controllers"
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
		upload, exists, err := artworkStore.Upload(ctx, input.Artwork.File, input.Artwork.Header.Filename)
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
