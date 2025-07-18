package models

import (
	"encoding/json"
	"fmt"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

type CdsSettings struct {
	Resource      string             `json:"resource"`
	ApiVersion    string             `json:"apiVersion"`
	TenantId      string             `json:"tenantId"`
	NamespaceId   string             `json:"namespaceId"`
	UseCommunity  bool               `json:"useCommunity"`
	CommunityId   string             `json:"communityId"`
	ClientId      string             `json:"clientId"`
	OauthPassThru bool               `json:"oauthPassThru"`
	Secrets       *SecretCdsSettings `json:"-"`
}

type SecretCdsSettings struct {
	ClientSecret string
}

func LoadPluginSettings(source backend.DataSourceInstanceSettings) (*CdsSettings, error) {
	settings := CdsSettings{}
	err := json.Unmarshal(source.JSONData, &settings)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal CdsSettings json: %w", err)
	}

	settings.Secrets = loadSecretPluginSettings(source.DecryptedSecureJSONData)

	return &settings, nil
}

func loadSecretPluginSettings(source map[string]string) *SecretCdsSettings {
	return &SecretCdsSettings{
		ClientSecret: source["clientSecret"],
	}
}
