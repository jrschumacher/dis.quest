package auth

// DiscoverPDS returns the PDS base URL for a given handle (Bluesky username).
// For Bluesky, this is always https://bsky.social. In the future, this could look up a handle in DNS or other registry.
func DiscoverPDS(_ string) (string, error) {
	// For now, always return Bluesky's PDS endpoint
	return "https://bsky.social", nil
}
