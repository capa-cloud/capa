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
	"encoding/json"

	"github.com/valyala/fasthttp"
)

const (
	jsonContentTypeHeader = "application/json"
	metadataPrefix        = "metadata."
)

type option = func(ctx *fasthttp.RequestCtx)

// responseWithJSON overrides the content-type with application/json.
func responseWithJSON(code int, obj []byte) option {
	return func(ctx *fasthttp.RequestCtx) {
		ctx.Response.SetStatusCode(code)
		ctx.Response.SetBody(obj)
		ctx.Response.Header.SetContentType(jsonContentTypeHeader)
	}
}

// responseWithError sets error code and jsonized error message.
func responseWithError(code int, resp ErrorResponse) option {
	b, _ := json.Marshal(&resp)
	return responseWithJSON(code, b)
}

// responseWithEmpty sets 204 status code.
func responseWithEmpty() option {
	return func(ctx *fasthttp.RequestCtx) {
		ctx.Response.SetBody(nil)
		ctx.Response.SetStatusCode(fasthttp.StatusNoContent)
	}
}

// setRespMetadata sets metadata headers.
func setRespMetadata(metadata map[string]string) option {
	return func(ctx *fasthttp.RequestCtx) {
		for k, v := range metadata {
			ctx.Response.Header.Set(metadataPrefix+k, v)
		}
	}
}

// with sets a default application/json content type if content type is not present.
func with(code int, obj []byte) option {
	return func(ctx *fasthttp.RequestCtx) {
		ctx.Response.SetStatusCode(code)
		ctx.Response.SetBody(obj)

		if len(ctx.Response.Header.ContentType()) == 0 {
			ctx.Response.Header.SetContentType(jsonContentTypeHeader)
		}
	}
}

func respond(ctx *fasthttp.RequestCtx, options ...option) {
	for _, option := range options {
		option(ctx)
	}
}
