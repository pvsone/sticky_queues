package file_processing

import (
	"context"
)

type StickyTaskQueue struct {
	MultiTaskQueue  string
	SingleTaskQueue string
}

func (q StickyTaskQueue) GetStickyTaskQueue(ctx context.Context) (StickyTaskQueue, error) {
	return q, nil
}
