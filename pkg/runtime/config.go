package runtime

import "encoding/json"

type AppConfig struct {
	AppId        string `json:"app_id"`
	CallbackPort int    `json:"callback_port"`
}

type CapaRuntimeConfig struct {
	AppManagement AppConfig                  `json:"app"`
	Extends       map[string]json.RawMessage `json:"extends,omitempty"` // extend config
}

func ParseRuntimeConfig(data json.RawMessage) (*CapaRuntimeConfig, error) {
	cfg := &CapaRuntimeConfig{}
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
