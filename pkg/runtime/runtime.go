package runtime

import "context"

// Create a structure to define the sidecar functionality.
type CapaRuntime struct {
	ctx    context.Context
	cancel context.CancelFunc
	// configs
	runtimeConfig *CapaRuntimeConfig
}

// NewCapaRuntime returns a new runtime with the given runtime config and global config.
func NewCapaRuntime(runtimeConfig *CapaRuntimeConfig) *CapaRuntime {
	ctx, cancel := context.WithCancel(context.Background())
	return &CapaRuntime{
		ctx:           ctx,
		cancel:        cancel,
		runtimeConfig: runtimeConfig,
	}
}
