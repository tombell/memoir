package artworkstore

// Upload is a data model for an uploaded artwork file to return in an HTTP
// response.
type Upload struct {
	Key string `json:"key"`
}
