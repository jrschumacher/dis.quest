// Package xrpc provides DID resolution for ATProtocol Personal Data Servers
package xrpc

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// DIDResolver resolves DIDs to Personal Data Server endpoints
type DIDResolver struct {
	httpClient *http.Client
}

// NewDIDResolver creates a new DID resolver
func NewDIDResolver() *DIDResolver {
	return &DIDResolver{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// ResolvePDS resolves a DID to its Personal Data Server endpoint
func (r *DIDResolver) ResolvePDS(did string) (string, error) {
	if strings.HasPrefix(did, "did:plc:") {
		return r.resolvePlcDID(did)
	}
	if strings.HasPrefix(did, "did:web:") {
		// Extract domain from did:web
		domain := strings.TrimPrefix(did, "did:web:")
		return fmt.Sprintf("https://%s", domain), nil
	}
	return "", fmt.Errorf("unsupported DID method: %s", did)
}

// resolvePlcDID resolves a did:plc DID to get the PDS endpoint
func (r *DIDResolver) resolvePlcDID(did string) (string, error) {
	// Query the PLC directory
	plcURL := fmt.Sprintf("https://plc.directory/%s", did)

	resp, err := r.httpClient.Get(plcURL)
	if err != nil {
		return "", fmt.Errorf("failed to resolve DID %s: %w", did, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to resolve DID %s: status %d", did, resp.StatusCode)
	}

	var didDoc struct {
		Service []struct {
			Type            string `json:"type"`
			ServiceEndpoint string `json:"serviceEndpoint"`
		} `json:"service"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&didDoc); err != nil {
		return "", fmt.Errorf("failed to decode DID document for %s: %w", did, err)
	}

	// Find the AtprotoPersonalDataServer service
	for _, service := range didDoc.Service {
		if service.Type == "AtprotoPersonalDataServer" {
			return service.ServiceEndpoint, nil
		}
	}

	return "", fmt.Errorf("no PDS endpoint found in DID document for %s", did)
}

// ATUriComponents represents parsed components of an AT URI
type ATUriComponents struct {
	DID        string
	Collection string
	RKey       string
}

// ParseATUri parses an AT URI into its components
func ParseATUri(uri string) (*ATUriComponents, error) {
	if !strings.HasPrefix(uri, "at://") {
		return nil, fmt.Errorf("invalid AT URI format: %s", uri)
	}

	// Remove at:// prefix and split
	parts := strings.Split(strings.TrimPrefix(uri, "at://"), "/")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid AT URI format: %s", uri)
	}

	components := &ATUriComponents{
		DID:        parts[0],
		Collection: parts[1],
	}

	if len(parts) >= 3 {
		components.RKey = parts[2]
	}

	return components, nil
}