package runtime

import (
	"group.rxcloud/capa/pkg/actors"
)

type (
	// runtimeOpts encapsulates the components to include in the runtime.
	runtimeOpts struct {
		actors []actors.Actors
	}

	// Option is a function that customizes the runtime.
	Option func(o *runtimeOpts)
)

// WithActors adds actor components to the runtime.
func WithActors(actors ...actors.Actors) Option {
	return func(o *runtimeOpts) {
		o.actors = append(o.actors, actors...)
	}
}
