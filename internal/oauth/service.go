package oauth

import (
	"context"
	"fmt"

	"github.com/jrschumacher/dis.quest/internal/auth"
	"github.com/jrschumacher/dis.quest/internal/config"
)

// Service provides OAuth functionality using the configured provider
type Service struct {
	provider Provider
	cfg      *config.Config
}

// NewService creates a new OAuth service with the configured provider
func NewService(cfg *config.Config) (*Service, error) {
	providerType, err := ParseProviderType(cfg.OAuthProvider)
	if err != nil {
		return nil, fmt.Errorf("invalid OAuth provider configuration: %w", err)
	}

	provider, err := NewProvider(providerType, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create OAuth provider: %w", err)
	}

	// Log which provider was successfully created
	fmt.Printf("OAuth service initialized with provider: %s\n", provider.GetProviderName())

	return &Service{
		provider: provider,
		cfg:      cfg,
	}, nil
}

// GetAuthURL generates the OAuth authorization URL with PKCE
func (s *Service) GetAuthURL(state, codeChallenge string) string {
	return s.provider.GetAuthURL(state, codeChallenge)
}

// ExchangeToken exchanges authorization code for access token with DPoP binding
func (s *Service) ExchangeToken(ctx context.Context, code, codeVerifier string) (*TokenResult, error) {
	return s.provider.ExchangeToken(ctx, code, codeVerifier)
}

// CreateAuthorizedClient creates an XRPC client with the given token
func (s *Service) CreateAuthorizedClient(token *TokenResult) (XRPCClient, error) {
	return s.provider.CreateAuthorizedClient(token)
}

// RefreshToken refreshes an expired access token
func (s *Service) RefreshToken(ctx context.Context, refreshToken string) (*TokenResult, error) {
	return s.provider.RefreshToken(ctx, refreshToken)
}

// GetProviderName returns the name of the active provider
func (s *Service) GetProviderName() string {
	return s.provider.GetProviderName()
}

// GeneratePKCE generates PKCE parameters for OAuth flow
func (s *Service) GeneratePKCE() (codeVerifier, codeChallenge string, err error) {
	return auth.GeneratePKCE()
}