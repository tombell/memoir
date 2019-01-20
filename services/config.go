package services

// Config contains any configuration data for the service functions.
type Config struct {
	AWS struct {
		Key    string
		Secret string
	}
}
