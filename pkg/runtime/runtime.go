package runtime

import (
	"context"
	log "github.com/sirupsen/logrus"
	"group.rxcloud/capa/pkg/actors"
	"group.rxcloud/capa/pkg/grpc"
	"os"
	"time"
)

type CapaRuntime struct {
	// context
	ctx       context.Context
	cancel    context.CancelFunc
	shutdownC chan error

	// configs
	runtimeConfig *CapaRuntimeConfig

	// actor
	actor actors.Actors
}

// NewCapaRuntime returns a new runtime with the given runtime config.
func NewCapaRuntime(runtimeConfig *CapaRuntimeConfig) *CapaRuntime {
	ctx, cancel := context.WithCancel(context.Background())
	return &CapaRuntime{
		ctx:           ctx,
		cancel:        cancel,
		runtimeConfig: runtimeConfig,
	}
}

// Run performs initialization of the runtime with the runtime and global configurations.
func (a *CapaRuntime) Run(opts ...Option) error {
	start := time.Now().UTC()

	appConfig := a.runtimeConfig.AppManagement
	log.Infof("[Capa.runtime.args] app id: %s", appConfig.AppId)
	log.Infof("[Capa.runtime.args] app env: %s", appConfig.Env)
	log.Infof("[Capa.runtime.args] app cloud: %s", appConfig.Cloud)
	sidecarConfig := a.runtimeConfig.SidecarManagement
	log.Infof("[Capa.runtime.args] runtime port: %d", sidecarConfig.RuntimePort)
	log.Infof("[Capa.runtime.args] runtime callback port: %d", sidecarConfig.RuntimeCallbackPort)
	log.Infof("[Capa.runtime.args] runtime shutdown duration: %s", sidecarConfig.GracefulShutdownDuration)

	// init options
	var o runtimeOpts
	for _, opt := range opts {
		opt(&o)
	}

	// init runtime
	err := a.initRuntime(&o)
	if err != nil {
		return err
	}

	duration := time.Since(start).Seconds() * 1000
	log.Infof("[Capa.runtime.Run] capa initialized. Status: Running. Init Elapsed %vms", duration)

	return nil
}

func (a *CapaRuntime) initRuntime(opts *runtimeOpts) error {
	// Create and start external gRPC servers
	grpcAPI := a.getGRPCAPI()

	err := a.startGRPCAPIServer(grpcAPI, a.runtimeConfig.SidecarManagement.RuntimePort)
	if err != nil {
		log.Fatalf("failed to start API gRPC server: %s", err)
	}

	//err = a.initActors()
	//if err != nil {
	//	log.Warnf("failed to init actors: %v", err)
	//} else {
	//	a.daprHTTPAPI.SetActorRuntime(a.actor)
	//	grpcAPI.SetActorRuntime(a.actor)
	//}
	return nil
}

func (a *CapaRuntime) getGRPCAPI() grpc.API {
	return grpc.NewAPI(a.runtimeConfig.AppManagement.AppId, a.actor, a.ShutdownWithWait)
}

func (a *CapaRuntime) startGRPCAPIServer(api grpc.API, port int) error {
	serverConf := a.getNewServerConfig(a.runtimeConfig.SidecarManagement.APIListenAddresses, port)
	server := grpc.NewAPIServer(api, serverConf)
	if err := server.StartNonBlocking(); err != nil {
		return err
	}
	return nil
}

func (a *CapaRuntime) getNewServerConfig(apiListenAddresses []string, port int) grpc.ServerConfig {
	return grpc.NewServerConfig(a.runtimeConfig.AppManagement.AppId, apiListenAddresses, port)
}

// ShutdownWithWait will gracefully stop runtime and wait outstanding operations.
func (a *CapaRuntime) ShutdownWithWait() {
	a.Shutdown(a.runtimeConfig.SidecarManagement.GracefulShutdownDuration)
	os.Exit(0)
}

func (a *CapaRuntime) Shutdown(duration time.Duration) {
	a.cancel()
	a.stopActor()
	log.Infof("dapr shutting down.")
	log.Info("Stopping Dapr APIs")
	log.Infof("Waiting %s to finish outstanding operations", duration)
	<-time.After(duration)
	a.shutdownC <- nil
}

func (a *CapaRuntime) stopActor() {
	if a.actor != nil {
		log.Info("Shutting down actor")
		//a.actor.Stop()
	}
}
