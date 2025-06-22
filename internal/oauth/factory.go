package oauth

import (
	"fmt"

	"github.com/jrschumacher/dis.quest/internal/config"
)

// ProviderType represents the different OAuth provider implementations
type ProviderType string

const (
	ProviderTypeManual  ProviderType = "manual"
	ProviderTypeTangled ProviderType = "tangled"
)

// NewProvider creates an OAuth provider based on the specified type
func NewProvider(providerType ProviderType, cfg *config.Config) (Provider, error) {
	oauthConfig := &Config{
		ClientID:       cfg.OAuthClientID,
		ClientURI:      cfg.PublicDomain,
		RedirectURI:    cfg.OAuthRedirectURL,
		PDSEndpoint:    cfg.PDSEndpoint,
		JWKSPrivateKey: cfg.JWKSPrivate,
		JWKSPublicKey:  cfg.JWKSPublic,
		Scope:          "atproto transition:generic",
	}

	switch providerType {
	case ProviderTypeManual:
		return NewManualProvider(oauthConfig), nil
	case ProviderTypeTangled:
		return NewTangledOAuthProvider(oauthConfig), nil
	default:
		return nil, fmt.Errorf("unknown OAuth provider type: %s", providerType)
	}
}

// ParseProviderType converts a string to ProviderType with validation
func ParseProviderType(s string) (ProviderType, error) {
	switch s {
	case string(ProviderTypeManual):
		return ProviderTypeManual, nil
	case string(ProviderTypeTangled):
		return ProviderTypeTangled, nil
	default:
		return "", fmt.Errorf("invalid provider type: %s (valid options: manual, tangled)", s)
	}
}