package grpc

import (
	"context"
	log "github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/emptypb"
	"group.rxcloud/capa/pkg/actors"
	runtimev1pb "group.rxcloud/capa/pkg/proto/runtime/v1"
	"sync"
)

// API is the gRPC interface for the Dapr gRPC API. It implements both the internal and external proto definitions.
type API interface {
	// Dapr Service methods
	SetActorRuntime(actor actors.Actors)
	RegisterActorTimer(ctx context.Context, in *runtimev1pb.RegisterActorTimerRequest) (*emptypb.Empty, error)
	UnregisterActorTimer(ctx context.Context, in *runtimev1pb.UnregisterActorTimerRequest) (*emptypb.Empty, error)
	RegisterActorReminder(ctx context.Context, in *runtimev1pb.RegisterActorReminderRequest) (*emptypb.Empty, error)
	UnregisterActorReminder(ctx context.Context, in *runtimev1pb.UnregisterActorReminderRequest) (*emptypb.Empty, error)
	RenameActorReminder(ctx context.Context, in *runtimev1pb.RenameActorReminderRequest) (*emptypb.Empty, error)
	GetActorState(ctx context.Context, in *runtimev1pb.GetActorStateRequest) (*runtimev1pb.GetActorStateResponse, error)
	ExecuteActorStateTransaction(ctx context.Context, in *runtimev1pb.ExecuteActorStateTransactionRequest) (*emptypb.Empty, error)
	InvokeActor(ctx context.Context, in *runtimev1pb.InvokeActorRequest) (*runtimev1pb.InvokeActorResponse, error)
	// Gets metadata of the sidecar
	GetMetadata(ctx context.Context, in *emptypb.Empty) (*runtimev1pb.GetMetadataResponse, error)
	// Sets value in extended metadata of the sidecar
	SetMetadata(ctx context.Context, in *runtimev1pb.SetMetadataRequest) (*emptypb.Empty, error)
	// Shutdown the sidecar
	Shutdown(ctx context.Context, in *emptypb.Empty) (*emptypb.Empty, error)
}

type api struct {
	id               string
	appProtocol      string
	extendedMetadata sync.Map

	actor actors.Actors

	shutdown func()
}

// NewAPI returns a new gRPC API.
func NewAPI(
	appID string,
	actor actors.Actors,
	shutdown func(),
) API {
	return &api{
		actor:    actor,
		id:       appID,
		shutdown: shutdown,
	}
}

func (a *api) RegisterActorTimer(ctx context.Context, in *runtimev1pb.RegisterActorTimerRequest) (*emptypb.Empty, error) {
	log.Infof("Actor")
	return &emptypb.Empty{}, nil
}

func (a *api) UnregisterActorTimer(ctx context.Context, in *runtimev1pb.UnregisterActorTimerRequest) (*emptypb.Empty, error) {
	log.Infof("Actor")
	return &emptypb.Empty{}, nil
}

func (a *api) RegisterActorReminder(ctx context.Context, in *runtimev1pb.RegisterActorReminderRequest) (*emptypb.Empty, error) {
	log.Infof("Actor")
	return &emptypb.Empty{}, nil
}

func (a *api) UnregisterActorReminder(ctx context.Context, in *runtimev1pb.UnregisterActorReminderRequest) (*emptypb.Empty, error) {
	log.Infof("Actor")
	return &emptypb.Empty{}, nil
}

func (a *api) RenameActorReminder(ctx context.Context, in *runtimev1pb.RenameActorReminderRequest) (*emptypb.Empty, error) {
	log.Infof("Actor")
	return &emptypb.Empty{}, nil
}

func (a *api) GetActorState(ctx context.Context, in *runtimev1pb.GetActorStateRequest) (*runtimev1pb.GetActorStateResponse, error) {
	log.Infof("Actor")
	return &runtimev1pb.GetActorStateResponse{}, nil
}

func (a *api) ExecuteActorStateTransaction(ctx context.Context, in *runtimev1pb.ExecuteActorStateTransactionRequest) (*emptypb.Empty, error) {
	log.Infof("Actor")
	return &emptypb.Empty{}, nil
}

func (a *api) InvokeActor(ctx context.Context, in *runtimev1pb.InvokeActorRequest) (*runtimev1pb.InvokeActorResponse, error) {
	log.Infof("Actor")
	return &runtimev1pb.InvokeActorResponse{}, nil
}

func (a *api) SetActorRuntime(actor actors.Actors) {
	a.actor = actor
}

func (a *api) GetMetadata(ctx context.Context, in *emptypb.Empty) (*runtimev1pb.GetMetadataResponse, error) {
	temp := make(map[string]string)

	// Copy synchronously so it can be serialized to JSON.
	a.extendedMetadata.Range(func(key, value interface{}) bool {
		temp[key.(string)] = value.(string)
		return true
	})
	response := &runtimev1pb.GetMetadataResponse{
		ExtendedMetadata: temp,
	}
	return response, nil
}

// SetMetadata Sets value in extended metadata of the sidecar.
func (a *api) SetMetadata(ctx context.Context, in *runtimev1pb.SetMetadataRequest) (*emptypb.Empty, error) {
	a.extendedMetadata.Store(in.Key, in.Value)
	return &emptypb.Empty{}, nil
}

// Shutdown the sidecar.
func (a *api) Shutdown(ctx context.Context, in *emptypb.Empty) (*emptypb.Empty, error) {
	go func() {
		<-ctx.Done()
		a.shutdown()
	}()
	return &emptypb.Empty{}, nil
}
