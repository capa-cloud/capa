package runtime

import (
	"encoding/json"
	"time"
)

type AppConfig struct {
	AppId string `json:"app_id"`
	Env   string `json:"env"`
	Cloud string `json:"cloud"`
}

type SidecarConfig struct {
	APIListenAddresses  []string `json:"api_listen_addresses"`
	RuntimePort         int      `json:"runtime_port"`
	RuntimeCallbackPort int      `json:"runtime_callback_port"`

	GracefulShutdownDuration time.Duration `json:"graceful_shutdown_duration"`
}

type CapaRuntimeConfig struct {
	AppManagement     *AppConfig     `json:"app"`
	SidecarManagement *SidecarConfig `json:"sidecar"`

	Extends map[string]json.RawMessage `json:"extends,omitempty"` // extend config
}

func ParseRuntimeConfig(data json.RawMessage) (*CapaRuntimeConfig, error) {
	cfg := &CapaRuntimeConfig{}
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
