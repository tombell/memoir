package task

// Task is the interface that wraps a piece of runnable functionality.
type Task interface {
	Name() string
	Run() error
}
