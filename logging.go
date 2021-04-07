package etu

import "github.com/icco/gutil/logging"

var (
	log = logging.Must(logging.NewLogger(Service))
)

const (
	// Service is the service this is deployed to in GCP.
	Service = "etu"
)
