package pds

import (
	"fmt"

	"github.com/jrschumacher/dis.quest/internal/config"
	"github.com/jrschumacher/dis.quest/internal/logger"
)

// ServiceFactory creates PDS services based on configuration
type ServiceFactory struct {
	config *config.Config
}

// NewServiceFactory creates a new PDS service factory
func NewServiceFactory(cfg *config.Config) *ServiceFactory {
	return &ServiceFactory{config: cfg}
}

// CreateATProtoService creates an ATProtocol service based on configuration
func (f *ServiceFactory) CreateATProtoService() ATProtoService {
	// For now, always return mock service for local development
	// In the future, this could switch based on config.AppEnv or specific PDS settings
	
	switch f.config.AppEnv {
	case config.EnvTest:
		logger.Info("Creating mock ATProto service for testing")
		return NewMockATProtoService()
	case config.EnvDev:
		logger.Info("Creating mock ATProto service for development")
		return NewMockATProtoService()
	case config.EnvProd:
		// In production, we would create a real ATProto service
		// For now, log and fall back to mock
		logger.Warn("Production ATProto service not yet implemented, using mock")
		return NewMockATProtoService()
	default:
		logger.Warn("Unknown environment, using mock ATProto service", "env", f.config.AppEnv)
		return NewMockATProtoService()
	}
}

// CreateLegacyService creates a legacy PDS service (for backward compatibility)
func (f *ServiceFactory) CreateLegacyService() Service {
	logger.Info("Creating legacy mock PDS service")
	return NewMockService()
}

// ServiceConfig holds PDS service configuration
type ServiceConfig struct {
	PDSEndpoint string
	UseOAuth    bool
	MockMode    bool
}

// GetServiceConfig extracts PDS configuration from app config
func (f *ServiceFactory) GetServiceConfig() *ServiceConfig {
	return &ServiceConfig{
		PDSEndpoint: f.config.PDSEndpoint,
		UseOAuth:    true, // Always use OAuth for security
		MockMode:    f.config.AppEnv != config.EnvProd,
	}
}

// ValidateConfig validates PDS service configuration
func (f *ServiceFactory) ValidateConfig() error {
	cfg := f.GetServiceConfig()
	
	if cfg.PDSEndpoint == "" {
		return fmt.Errorf("PDS endpoint is required")
	}
	
	if !cfg.MockMode && cfg.PDSEndpoint == "http://localhost:4000" {
		logger.Warn("Using localhost PDS endpoint in non-mock mode - ensure PDS is running")
	}
	
	return nil
}