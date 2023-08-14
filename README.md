The Workflow has 3 Activities
* Download File
* Process File
* Delete File

The activities use sticky queues in order to ensure they run on the same host.

The Download and Delete activities can be run concurrently for multiple files on the same host. However the Process activity must run sequentially across all of the files on the same host.

This is achieved by creating separate Worker entities for Download/Delete vs. Process.  The "Process" Worker entity is configured with a `MaxConcurrentActivityExecutionSize` of 1.
