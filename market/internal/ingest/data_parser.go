package ingest

import (
	"context"
	"log"
)

// Request represents an ingest payload coming from the control queue.
type IngestRequest struct {
	FileName string `json:"file_name"`
}

// Handle processes an ingest request.
func HandleIngest(ctx context.Context, request IngestRequest) {
	log.Printf("Handing request to ingest data: %v", request)
}
