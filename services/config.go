package services

// AWSConfig contains credentials for using AWS services.
type AWSConfig struct {
	Key    string
	Secret string
}

// Config contains any configuration data for the service functions.
type Config struct {
	AWS *AWSConfig
}
