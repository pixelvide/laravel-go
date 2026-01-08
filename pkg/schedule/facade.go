package schedule

import (
	"sync"
)

var (
	globalKernel *Kernel
	once         sync.Once
)

// SetGlobalKernel sets the global kernel instance
func SetGlobalKernel(k *Kernel) {
	globalKernel = k
}

// GetGlobalKernel returns the global kernel, initializing a default one if needed
func GetGlobalKernel() *Kernel {
	once.Do(func() {
		if globalKernel == nil {
			globalKernel = NewKernel(nil)
		}
	})
	return globalKernel
}

// Register adds a task to the global scheduler
func Register(schedule string, cmd func(), opts ...JobOption) {
	GetGlobalKernel().Register(schedule, cmd, opts...)
}
