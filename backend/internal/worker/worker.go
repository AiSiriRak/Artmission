package worker

import "context"

// Worker is a background job that runs until ctx is cancelled.
type Worker interface {
	Run(ctx context.Context)
}
