/*
Copyright 2021 The Dapr Authors
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package http

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/dapr/dapr/pkg/actors"
	"github.com/dapr/dapr/pkg/channel/http"
	"github.com/dapr/dapr/pkg/messages"
	invokev1 "github.com/dapr/dapr/pkg/messaging/v1"
	"github.com/valyala/fasthttp"
	"google.golang.org/grpc/codes"
	"strconv"
	"strings"
	"sync"
)

// API returns a list of HTTP endpoints for Capa.
type API interface {
	MarkStatusAsReady()

	APIEndpoints() []Endpoint
	PublicEndpoints() []Endpoint

	SetActorRuntime(actor actors.Actors)
}

type api struct {
	id               string
	readyStatus      bool
	extendedMetadata sync.Map
	shutdown         func()

	endpoints       []Endpoint
	publicEndpoints []Endpoint

	actor actors.Actors
}

type metadata struct {
	ID                string                     `json:"id"`
	ActiveActorsCount []actors.ActiveActorsCount `json:"actors"`
	Extended          map[string]string          `json:"extended"`
}

const (
	apiVersionV1   = "v1.0"
	methodParam    = "method"
	actorTypeParam = "actorType"
	actorIDParam   = "actorId"
	nameParam      = "name"
	daprAppID      = "dapr-app-id"
)

// NewAPI returns a new API.
func NewAPI(
	appID string,
	actor actors.Actors,
	shutdown func(),
) API {
	api := &api{
		actor:    actor,
		id:       appID,
		shutdown: shutdown,
	}

	metadataEndpoints := api.constructMetadataEndpoints()
	healthEndpoints := api.constructHealthzEndpoints()
	api.publicEndpoints = append(api.publicEndpoints, metadataEndpoints...)
	api.publicEndpoints = append(api.publicEndpoints, healthEndpoints...)

	actorEndpoints := api.constructActorEndpoints()
	shutdownEndpoints := api.constructShutdownEndpoints()
	api.endpoints = append(api.endpoints, actorEndpoints...)
	api.endpoints = append(api.endpoints, metadataEndpoints...)
	api.endpoints = append(api.endpoints, shutdownEndpoints...)
	api.endpoints = append(api.endpoints, healthEndpoints...)

	return api
}

// APIEndpoints returns the list of registered endpoints.
func (a *api) APIEndpoints() []Endpoint {
	return a.endpoints
}

// PublicEndpoints returns the list of registered endpoints.
func (a *api) PublicEndpoints() []Endpoint {
	return a.publicEndpoints
}

// MarkStatusAsReady marks the ready status of dapr.
func (a *api) MarkStatusAsReady() {
	a.readyStatus = true
}

func (a *api) constructActorEndpoints() []Endpoint {
	return []Endpoint{
		{
			Methods: []string{fasthttp.MethodPost, fasthttp.MethodPut},
			Route:   "actors/{actorType}/{actorId}/state",
			Version: apiVersionV1,
			Handler: a.onActorStateTransaction,
		},
		{
			Methods: []string{fasthttp.MethodGet, fasthttp.MethodPost, fasthttp.MethodDelete, fasthttp.MethodPut},
			Route:   "actors/{actorType}/{actorId}/method/{method}",
			Version: apiVersionV1,
			Handler: a.onDirectActorMessage,
		},
		{
			Methods: []string{fasthttp.MethodGet},
			Route:   "actors/{actorType}/{actorId}/state/{key}",
			Version: apiVersionV1,
			Handler: a.onGetActorState,
		},
		{
			Methods: []string{fasthttp.MethodPost, fasthttp.MethodPut},
			Route:   "actors/{actorType}/{actorId}/reminders/{name}",
			Version: apiVersionV1,
			Handler: a.onCreateActorReminder,
		},
		{
			Methods: []string{fasthttp.MethodPost, fasthttp.MethodPut},
			Route:   "actors/{actorType}/{actorId}/timers/{name}",
			Version: apiVersionV1,
			Handler: a.onCreateActorTimer,
		},
		{
			Methods: []string{fasthttp.MethodDelete},
			Route:   "actors/{actorType}/{actorId}/reminders/{name}",
			Version: apiVersionV1,
			Handler: a.onDeleteActorReminder,
		},
		{
			Methods: []string{fasthttp.MethodDelete},
			Route:   "actors/{actorType}/{actorId}/timers/{name}",
			Version: apiVersionV1,
			Handler: a.onDeleteActorTimer,
		},
		{
			Methods: []string{fasthttp.MethodGet},
			Route:   "actors/{actorType}/{actorId}/reminders/{name}",
			Version: apiVersionV1,
			Handler: a.onGetActorReminder,
		},
		{
			Methods: []string{fasthttp.MethodPatch},
			Route:   "actors/{actorType}/{actorId}/reminders/{name}",
			Version: apiVersionV1,
			Handler: a.onRenameActorReminder,
		},
	}
}

func (a *api) constructMetadataEndpoints() []Endpoint {
	return []Endpoint{
		{
			Methods: []string{fasthttp.MethodGet},
			Route:   "metadata",
			Version: apiVersionV1,
			Handler: a.onGetMetadata,
		},
		{
			Methods: []string{fasthttp.MethodPut},
			Route:   "metadata/{key}",
			Version: apiVersionV1,
			Handler: a.onPutMetadata,
		},
	}
}

func (a *api) constructShutdownEndpoints() []Endpoint {
	return []Endpoint{
		{
			Methods: []string{fasthttp.MethodPost},
			Route:   "shutdown",
			Version: apiVersionV1,
			Handler: a.onShutdown,
		},
	}
}

func (a *api) constructHealthzEndpoints() []Endpoint {
	return []Endpoint{
		{
			Methods: []string{fasthttp.MethodGet},
			Route:   "healthz",
			Version: apiVersionV1,
			Handler: a.onGetHealthz,
		},
	}
}

func (a *api) onCreateActorReminder(reqCtx *fasthttp.RequestCtx) {
	if a.actor == nil {
		msg := NewErrorResponse("ERR_ACTOR_RUNTIME_NOT_FOUND", messages.ErrActorRuntimeNotFound)
		respond(reqCtx, responseWithError(fasthttp.StatusInternalServerError, msg))
		return
	}

	actorType := reqCtx.UserValue(actorTypeParam).(string)
	actorID := reqCtx.UserValue(actorIDParam).(string)
	name := reqCtx.UserValue(nameParam).(string)

	var req actors.CreateReminderRequest
	err := json.Unmarshal(reqCtx.PostBody(), &req)
	if err != nil {
		msg := NewErrorResponse("ERR_MALFORMED_REQUEST", fmt.Sprintf(messages.ErrMalformedRequest, err))
		respond(reqCtx, responseWithError(fasthttp.StatusBadRequest, msg))
		log.Debug(msg)
		return
	}

	req.Name = name
	req.ActorType = actorType
	req.ActorID = actorID

	err = a.actor.CreateReminder(reqCtx, &req)
	if err != nil {
		msg := NewErrorResponse("ERR_ACTOR_REMINDER_CREATE", fmt.Sprintf(messages.ErrActorReminderCreate, err))
		respond(reqCtx, responseWithError(fasthttp.StatusInternalServerError, msg))
		log.Debug(msg)
	} else {
		respond(reqCtx, responseWithEmpty())
	}
}

func (a *api) onRenameActorReminder(reqCtx *fasthttp.RequestCtx) {
	if a.actor == nil {
		msg := NewErrorResponse("ERR_ACTOR_RUNTIME_NOT_FOUND", messages.ErrActorRuntimeNotFound)
		respond(reqCtx, responseWithError(fasthttp.StatusInternalServerError, msg))
		return
	}

	actorType := reqCtx.UserValue(actorTypeParam).(string)
	actorID := reqCtx.UserValue(actorIDParam).(string)
	name := reqCtx.UserValue(nameParam).(string)

	var req actors.RenameReminderRequest
	err := json.Unmarshal(reqCtx.PostBody(), &req)
	if err != nil {
		msg := NewErrorResponse("ERR_MALFORMED_REQUEST", fmt.Sprintf(messages.ErrMalformedRequest, err))
		respond(reqCtx, responseWithError(fasthttp.StatusBadRequest, msg))
		log.Debug(msg)
		return
	}

	req.OldName = name
	req.ActorType = actorType
	req.ActorID = actorID

	err = a.actor.RenameReminder(reqCtx, &req)
	if err != nil {
		msg := NewErrorResponse("ERR_ACTOR_REMINDER_RENAME", fmt.Sprintf(messages.ErrActorReminderRename, err))
		respond(reqCtx, responseWithError(fasthttp.StatusInternalServerError, msg))
		log.Debug(msg)
	} else {
		respond(reqCtx, responseWithEmpty())
	}
}

func (a *api) onCreateActorTimer(reqCtx *fasthttp.RequestCtx) {
	if a.actor == nil {
		msg := NewErrorResponse("ERR_ACTOR_RUNTIME_NOT_FOUND", messages.ErrActorRuntimeNotFound)
		respond(reqCtx, responseWithError(fasthttp.StatusInternalServerError, msg))
		log.Debug(msg)
		return
	}

	actorType := reqCtx.UserValue(actorTypeParam).(string)
	actorID := reqCtx.UserValue(actorIDParam).(string)
	name := reqCtx.UserValue(nameParam).(string)

	var req actors.CreateTimerRequest
	err := json.Unmarshal(reqCtx.PostBody(), &req)
	if err != nil {
		msg := NewErrorResponse("ERR_MALFORMED_REQUEST", fmt.Sprintf(messages.ErrMalformedRequest, err))
		respond(reqCtx, responseWithError(fasthttp.StatusBadRequest, msg))
		log.Debug(msg)
		return
	}

	req.Name = name
	req.ActorType = actorType
	req.ActorID = actorID

	err = a.actor.CreateTimer(reqCtx, &req)
	if err != nil {
		msg := NewErrorResponse("ERR_ACTOR_TIMER_CREATE", fmt.Sprintf(messages.ErrActorTimerCreate, err))
		respond(reqCtx, responseWithError(fasthttp.StatusInternalServerError, msg))
		log.Debug(msg)
	} else {
		respond(reqCtx, responseWithEmpty())
	}
}

func (a *api) onDeleteActorReminder(reqCtx *fasthttp.RequestCtx) {
	if a.actor == nil {
		msg := NewErrorResponse("ERR_ACTOR_RUNTIME_NOT_FOUND", messages.ErrActorRuntimeNotFound)
		respond(reqCtx, responseWithError(fasthttp.StatusInternalServerError, msg))
		log.Debug(msg)
		return
	}

	actorType := reqCtx.UserValue(actorTypeParam).(string)
	actorID := reqCtx.UserValue(actorIDParam).(string)
	name := reqCtx.UserValue(nameParam).(string)

	req := actors.DeleteReminderRequest{
		Name:      name,
		ActorID:   actorID,
		ActorType: actorType,
	}

	err := a.actor.DeleteReminder(reqCtx, &req)
	if err != nil {
		msg := NewErrorResponse("ERR_ACTOR_REMINDER_DELETE", fmt.Sprintf(messages.ErrActorReminderDelete, err))
		respond(reqCtx, responseWithError(fasthttp.StatusInternalServerError, msg))
		log.Debug(msg)
	} else {
		respond(reqCtx, responseWithEmpty())
	}
}

func (a *api) onActorStateTransaction(reqCtx *fasthttp.RequestCtx) {
	if a.actor == nil {
		msg := NewErrorResponse("ERR_ACTOR_RUNTIME_NOT_FOUND", messages.ErrActorRuntimeNotFound)
		respond(reqCtx, responseWithError(fasthttp.StatusInternalServerError, msg))
		log.Debug(msg)
		return
	}

	actorType := reqCtx.UserValue(actorTypeParam).(string)
	actorID := reqCtx.UserValue(actorIDParam).(string)
	body := reqCtx.PostBody()

	var ops []actors.TransactionalOperation
	err := json.Unmarshal(body, &ops)
	if err != nil {
		msg := NewErrorResponse("ERR_MALFORMED_REQUEST", err.Error())
		respond(reqCtx, responseWithError(fasthttp.StatusBadRequest, msg))
		log.Debug(msg)
		return
	}

	hosted := a.actor.IsActorHosted(reqCtx, &actors.ActorHostedRequest{
		ActorType: actorType,
		ActorID:   actorID,
	})

	if !hosted {
		msg := NewErrorResponse("ERR_ACTOR_INSTANCE_MISSING", messages.ErrActorInstanceMissing)
		respond(reqCtx, responseWithError(fasthttp.StatusBadRequest, msg))
		log.Debug(msg)
		return
	}

	req := actors.TransactionalRequest{
		ActorID:    actorID,
		ActorType:  actorType,
		Operations: ops,
	}

	err = a.actor.TransactionalStateOperation(reqCtx, &req)
	if err != nil {
		msg := NewErrorResponse("ERR_ACTOR_STATE_TRANSACTION_SAVE", fmt.Sprintf(messages.ErrActorStateTransactionSave, err))
		respond(reqCtx, responseWithError(fasthttp.StatusInternalServerError, msg))
		log.Debug(msg)
	} else {
		respond(reqCtx, responseWithEmpty())
	}
}

func (a *api) onGetActorReminder(reqCtx *fasthttp.RequestCtx) {
	if a.actor == nil {
		msg := NewErrorResponse("ERR_ACTOR_RUNTIME_NOT_FOUND", messages.ErrActorRuntimeNotFound)
		respond(reqCtx, responseWithError(fasthttp.StatusInternalServerError, msg))
		log.Debug(msg)
		return
	}

	actorType := reqCtx.UserValue(actorTypeParam).(string)
	actorID := reqCtx.UserValue(actorIDParam).(string)
	name := reqCtx.UserValue(nameParam).(string)

	resp, err := a.actor.GetReminder(reqCtx, &actors.GetReminderRequest{
		ActorType: actorType,
		ActorID:   actorID,
		Name:      name,
	})
	if err != nil {
		msg := NewErrorResponse("ERR_ACTOR_REMINDER_GET", fmt.Sprintf(messages.ErrActorReminderGet, err))
		respond(reqCtx, responseWithError(fasthttp.StatusInternalServerError, msg))
		log.Debug(msg)
		return
	}
	b, err := json.Marshal(resp)
	if err != nil {
		msg := NewErrorResponse("ERR_ACTOR_REMINDER_GET", fmt.Sprintf(messages.ErrActorReminderGet, err))
		respond(reqCtx, responseWithError(fasthttp.StatusInternalServerError, msg))
		log.Debug(msg)
		return
	}

	respond(reqCtx, responseWithJSON(fasthttp.StatusOK, b))
}

func (a *api) onDeleteActorTimer(reqCtx *fasthttp.RequestCtx) {
	if a.actor == nil {
		msg := NewErrorResponse("ERR_ACTOR_RUNTIME_NOT_FOUND", messages.ErrActorRuntimeNotFound)
		respond(reqCtx, responseWithError(fasthttp.StatusInternalServerError, msg))
		log.Debug(msg)
		return
	}

	actorType := reqCtx.UserValue(actorTypeParam).(string)
	actorID := reqCtx.UserValue(actorIDParam).(string)
	name := reqCtx.UserValue(nameParam).(string)

	req := actors.DeleteTimerRequest{
		Name:      name,
		ActorID:   actorID,
		ActorType: actorType,
	}
	err := a.actor.DeleteTimer(reqCtx, &req)
	if err != nil {
		msg := NewErrorResponse("ERR_ACTOR_TIMER_DELETE", fmt.Sprintf(messages.ErrActorTimerDelete, err))
		respond(reqCtx, responseWithError(fasthttp.StatusInternalServerError, msg))
		log.Debug(msg)
	} else {
		respond(reqCtx, responseWithEmpty())
	}
}

func (a *api) onDirectActorMessage(reqCtx *fasthttp.RequestCtx) {
	if a.actor == nil {
		msg := NewErrorResponse("ERR_ACTOR_RUNTIME_NOT_FOUND", messages.ErrActorRuntimeNotFound)
		respond(reqCtx, responseWithError(fasthttp.StatusInternalServerError, msg))
		log.Debug(msg)
		return
	}

	actorType := reqCtx.UserValue(actorTypeParam).(string)
	actorID := reqCtx.UserValue(actorIDParam).(string)
	verb := strings.ToUpper(string(reqCtx.Method()))
	method := reqCtx.UserValue(methodParam).(string)
	body := reqCtx.PostBody()

	req := invokev1.NewInvokeMethodRequest(method)
	req.WithActor(actorType, actorID)
	req.WithHTTPExtension(verb, reqCtx.QueryArgs().String())
	req.WithRawData(body, string(reqCtx.Request.Header.ContentType()))

	// Save headers to metadata.
	metadata := map[string][]string{}
	reqCtx.Request.Header.VisitAll(func(key []byte, value []byte) {
		metadata[string(key)] = []string{string(value)}
	})
	req.WithMetadata(metadata)

	// Unlike other actor calls, resiliency is handled here for invocation.
	// This is due to actor invocation involving a lookup for the host.
	// Having the retry here allows us to capture that and be resilient to host failure.
	// Additionally, we don't perform timeouts at this level. This is because an actor
	// should technically wait forever on the locking mechanism. If we timeout while
	// waiting for the lock, we can also create a queue of calls that will try and continue
	// after the timeout.
	policy := a.resiliency.ActorPreLockPolicy(reqCtx, actorType, actorID)
	var resp *invokev1.InvokeMethodResponse
	err := policy(func(ctx context.Context) (rErr error) {
		resp, rErr = a.actor.Call(ctx, req)
		return rErr
	})
	if err != nil {
		msg := NewErrorResponse("ERR_ACTOR_INVOKE_METHOD", fmt.Sprintf(messages.ErrActorInvoke, err))
		respond(reqCtx, responseWithError(fasthttp.StatusInternalServerError, msg))
		log.Debug(msg)
		return
	}

	invokev1.InternalMetadataToHTTPHeader(reqCtx, resp.Headers(), reqCtx.Response.Header.Set)
	contentType, body := resp.RawData()
	reqCtx.Response.Header.SetContentType(contentType)

	// Construct response.
	statusCode := int(resp.Status().Code)
	if !resp.IsHTTPResponse() {
		statusCode = invokev1.HTTPStatusFromCode(codes.Code(statusCode))
	}
	respond(reqCtx, with(statusCode, body))
}

func (a *api) onGetActorState(reqCtx *fasthttp.RequestCtx) {
	if a.actor == nil {
		msg := NewErrorResponse("ERR_ACTOR_RUNTIME_NOT_FOUND", messages.ErrActorRuntimeNotFound)
		respond(reqCtx, responseWithError(fasthttp.StatusInternalServerError, msg))
		log.Debug(msg)
		return
	}

	actorType := reqCtx.UserValue(actorTypeParam).(string)
	actorID := reqCtx.UserValue(actorIDParam).(string)
	key := reqCtx.UserValue(stateKeyParam).(string)

	hosted := a.actor.IsActorHosted(reqCtx, &actors.ActorHostedRequest{
		ActorType: actorType,
		ActorID:   actorID,
	})

	if !hosted {
		msg := NewErrorResponse("ERR_ACTOR_INSTANCE_MISSING", messages.ErrActorInstanceMissing)
		respond(reqCtx, responseWithError(fasthttp.StatusBadRequest, msg))
		log.Debug(msg)
		return
	}

	req := actors.GetStateRequest{
		ActorType: actorType,
		ActorID:   actorID,
		Key:       key,
	}

	resp, err := a.actor.GetState(reqCtx, &req)
	if err != nil {
		msg := NewErrorResponse("ERR_ACTOR_STATE_GET", fmt.Sprintf(messages.ErrActorStateGet, err))
		respond(reqCtx, responseWithError(fasthttp.StatusInternalServerError, msg))
		log.Debug(msg)
	} else {
		if resp == nil || resp.Data == nil {
			respond(reqCtx, responseWithEmpty())
			return
		}
		respond(reqCtx, responseWithJSON(fasthttp.StatusOK, resp.Data))
	}
}

func (a *api) onGetMetadata(reqCtx *fasthttp.RequestCtx) {
	temp := make(map[string]string)

	// Copy synchronously so it can be serialized to JSON.
	a.extendedMetadata.Range(func(key, value interface{}) bool {
		temp[key.(string)] = key.(string)

		return true
	})

	activeActorsCount := []actors.ActiveActorsCount{}
	if a.actor != nil {
		activeActorsCount = a.actor.GetActiveActorsCount(reqCtx)
	}

	components := a.getComponentsFn()
	registeredComponents := make([]registeredComponent, 0, len(components))

	for _, comp := range components {
		registeredComp := registeredComponent{
			Name:    comp.Name,
			Version: comp.Spec.Version,
			Type:    comp.Spec.Type,
		}
		registeredComponents = append(registeredComponents, registeredComp)
	}

	mtd := metadata{
		ID:                   a.id,
		ActiveActorsCount:    activeActorsCount,
		Extended:             temp,
		RegisteredComponents: registeredComponents,
	}

	mtdBytes, err := json.Marshal(mtd)
	if err != nil {
		msg := NewErrorResponse("ERR_METADATA_GET", fmt.Sprintf(messages.ErrMetadataGet, err))
		respond(reqCtx, responseWithError(fasthttp.StatusInternalServerError, msg))
		log.Debug(msg)
	} else {
		respond(reqCtx, responseWithJSON(fasthttp.StatusOK, mtdBytes))
	}
}

func (a *api) onPutMetadata(reqCtx *fasthttp.RequestCtx) {
	key := fmt.Sprintf("%v", reqCtx.UserValue("key"))
	body := reqCtx.PostBody()
	a.extendedMetadata.Store(key, string(body))
	respond(reqCtx, responseWithEmpty())
}

func (a *api) onShutdown(reqCtx *fasthttp.RequestCtx) {
	if !reqCtx.IsPost() {
		log.Warn("Please use POST method when invoking shutdown API")
	}

	respond(reqCtx, responseWithEmpty())
	go func() {
		a.shutdown()
	}()
}

// GetStatusCodeFromMetadata extracts the http status code from the metadata if it exists.
func GetStatusCodeFromMetadata(metadata map[string]string) int {
	code := metadata[http.HTTPStatusCode]
	if code != "" {
		statusCode, err := strconv.Atoi(code)
		if err == nil {
			return statusCode
		}
	}

	return fasthttp.StatusOK
}

func (a *api) onGetHealthz(reqCtx *fasthttp.RequestCtx) {
	if !a.readyStatus {
		msg := NewErrorResponse("ERR_HEALTH_NOT_READY", messages.ErrHealthNotReady)
		respond(reqCtx, responseWithError(fasthttp.StatusInternalServerError, msg))
		log.Debug(msg)
	} else {
		respond(reqCtx, responseWithEmpty())
	}
}

func getMetadataFromRequest(reqCtx *fasthttp.RequestCtx) map[string]string {
	metadata := map[string]string{}
	reqCtx.QueryArgs().VisitAll(func(key []byte, value []byte) {
		queryKey := string(key)
		if strings.HasPrefix(queryKey, metadataPrefix) {
			k := strings.TrimPrefix(queryKey, metadataPrefix)
			metadata[k] = string(value)
		}
	})

	return metadata
}

func (a *api) SetActorRuntime(actor actors.Actors) {
	a.actor = actor
}
