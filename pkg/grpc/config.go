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

package grpc

// ServerConfig is the config object for a grpc server.
type ServerConfig struct {
	AppID              string
	APIListenAddresses []string
	Port               int
}

// NewServerConfig returns a new grpc server config.
func NewServerConfig(
	appID string,
	apiListenAddresses []string,
	port int) ServerConfig {
	return ServerConfig{
		AppID:              appID,
		APIListenAddresses: apiListenAddresses,
		Port:               port,
	}
}
