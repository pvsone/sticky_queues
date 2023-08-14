package file_processing

import (
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"go.temporal.io/sdk/workflow"
)

// FileProcessingWorkflow is a workflow that uses stick activity queues to process files
// on a consistent host.
func FileProcessingWorkflow(ctx workflow.Context) (err error) {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)
	var stickyTaskQueue StickyTaskQueue
	err = workflow.ExecuteActivity(ctx, "GetStickyTaskQueue").Get(ctx, &stickyTaskQueue)
	if err != nil {
		return
	}

	// Download and Delete will run on the multi-task queue
	mao := workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute,
		TaskQueue:           stickyTaskQueue.MultiTaskQueue,
	}
	mctx := workflow.WithActivityOptions(ctx, mao)

	downloadPath := filepath.Join("/tmp", uuid.New().String())
	err = workflow.ExecuteActivity(mctx, DownloadFile, "https://temporal.io", downloadPath).Get(mctx, nil)
	if err != nil {
		return
	}
	defer func() {
		err = workflow.ExecuteActivity(mctx, DeleteFile, downloadPath).Get(mctx, nil)
	}()

	// Process will run on the single-task queue
	sao := workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute,
		TaskQueue:           stickyTaskQueue.SingleTaskQueue,
	}
	sctx := workflow.WithActivityOptions(ctx, sao)

	err = workflow.ExecuteActivity(sctx, ProcessFile, downloadPath).Get(sctx, nil)
	return
}
