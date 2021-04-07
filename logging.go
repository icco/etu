package etu

var (
	log = logging.Must(logging.NewLogger(Service))
)

const (
	// Service is the service this is deployed to in GCP.
	Service = "etu"
)
