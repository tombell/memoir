package task

// Task is the interface that wraps a piece of runnable functionality.
type Task interface {
	// Name returns the name of the task.
	Name() string

	// Run executes the task.
	Run() error
}
