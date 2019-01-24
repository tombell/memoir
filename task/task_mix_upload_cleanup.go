package task

// MixUploadCleanupTask ...
type MixUploadCleanupTask struct{}

// Name returns the name of the task.
func (t *MixUploadCleanupTask) Name() string {
	return "mix-upload-cleanup"
}

// Run runs the task to cleanup any unassociated uploaded mixes.
func (t *MixUploadCleanupTask) Run() error {
	// TODO
	return nil
}
