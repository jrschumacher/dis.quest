package session

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jrschumacher/dis.quest/pkg/atproto/oauth"
)

// FileStorage implements Storage interface using local filesystem.
// This is suitable for CLI applications, desktop apps, and development.
// Sessions persist across application restarts.
type FileStorage struct {
	baseDir       string
	encryptionKey []byte
}

// NewFileStorage creates a new file-based session storage.
func NewFileStorage(baseDir string, encryptionKey []byte) Storage {
	return &FileStorage{
		baseDir:       baseDir,
		encryptionKey: encryptionKey,
	}
}

// Store saves session data to a file.
func (f *FileStorage) Store(ctx context.Context, key string, data *Data) error {
	if key == "" {
		return fmt.Errorf("session key cannot be empty")
	}
	if data == nil {
		return fmt.Errorf("session data cannot be nil")
	}

	// Ensure base directory exists
	if err := os.MkdirAll(f.baseDir, 0700); err != nil {
		return fmt.Errorf("failed to create session directory: %w", err)
	}

	// Create a copy of data for file storage
	fileData := *data

	// Encrypt DPoP key if encryption is configured
	if data.DPoPKey != nil && len(f.encryptionKey) > 0 {
		encrypted, err := f.encryptDPoPKey(data.DPoPKey)
		if err != nil {
			return fmt.Errorf("failed to encrypt DPoP key: %w", err)
		}
		fileData.DPoPKeyEncrypted = encrypted
		fileData.DPoPKey = nil // Don't store raw key in file
	}

	// Serialize session data
	jsonData, err := json.MarshalIndent(&fileData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize session data: %w", err)
	}

	// Write to file
	filePath := f.getFilePath(key)
	if err := os.WriteFile(filePath, jsonData, 0600); err != nil {
		return fmt.Errorf("failed to write session file: %w", err)
	}

	return nil
}

// Load retrieves session data from a file.
func (f *FileStorage) Load(ctx context.Context, key string) (*Data, error) {
	if key == "" {
		return nil, fmt.Errorf("session key cannot be empty")
	}

	filePath := f.getFilePath(key)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("session not found")
	}

	// Read file
	jsonData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read session file: %w", err)
	}

	// Deserialize session data
	var data Data
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return nil, fmt.Errorf("failed to deserialize session data: %w", err)
	}

	// Check if session has expired
	if time.Now().After(data.ExpiresAt) {
		return nil, fmt.Errorf("session expired")
	}

	// Decrypt DPoP key if present and encryption is configured
	if len(data.DPoPKeyEncrypted) > 0 && len(f.encryptionKey) > 0 {
		dpopKey, err := f.decryptDPoPKey(data.DPoPKeyEncrypted)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt DPoP key: %w", err)
		}
		data.DPoPKey = dpopKey
	}

	return &data, nil
}

// Delete removes session file.
func (f *FileStorage) Delete(ctx context.Context, key string) error {
	if key == "" {
		return fmt.Errorf("session key cannot be empty")
	}

	filePath := f.getFilePath(key)

	// Remove file (ignore if it doesn't exist)
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove session file: %w", err)
	}

	return nil
}

// Cleanup removes expired session files.
func (f *FileStorage) Cleanup(ctx context.Context) error {
	// Check if directory exists
	if _, err := os.Stat(f.baseDir); os.IsNotExist(err) {
		return nil // Nothing to clean up
	}

	// Walk through all session files
	return filepath.WalkDir(f.baseDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if d.IsDir() {
			return nil
		}

		// Skip non-session files
		if !strings.HasSuffix(d.Name(), ".json") {
			return nil
		}

		// Try to load session to check expiration
		sessionKey := strings.TrimSuffix(d.Name(), ".json")
		sessionKey = f.sanitizeKey(sessionKey)

		// Read and check expiration without full deserialization
		jsonData, err := os.ReadFile(path)
		if err != nil {
			// If we can't read the file, skip it
			return nil
		}

		var partial struct {
			ExpiresAt time.Time `json:"expires_at"`
		}
		if err := json.Unmarshal(jsonData, &partial); err != nil {
			// If we can't parse the file, skip it
			return nil
		}

		// Remove if expired
		if time.Now().After(partial.ExpiresAt) {
			if err := os.Remove(path); err != nil {
				// Log error but continue cleanup
				fmt.Printf("Failed to remove expired session file %s: %v\n", path, err)
			}
		}

		return nil
	})
}

// Close cleans up storage resources.
func (f *FileStorage) Close() error {
	// No persistent connections to close for file storage
	return nil
}

// getFilePath returns the full file path for a session key.
func (f *FileStorage) getFilePath(key string) string {
	filename := f.sanitizeKey(key) + ".json"
	return filepath.Join(f.baseDir, filename)
}

// sanitizeKey removes potentially dangerous characters from session keys.
func (f *FileStorage) sanitizeKey(key string) string {
	// Replace any path separators and other dangerous characters
	sanitized := strings.ReplaceAll(key, "/", "_")
	sanitized = strings.ReplaceAll(sanitized, "\\", "_")
	sanitized = strings.ReplaceAll(sanitized, "..", "_")
	return sanitized
}

// encryptDPoPKey encrypts a DPoP private key using the configured encryption key.
func (f *FileStorage) encryptDPoPKey(key *ecdsa.PrivateKey) ([]byte, error) {
	// Use the same encryption logic as the manager
	keyPair := &oauth.DPoPKeyPair{PrivateKey: key}
	pemData, err := keyPair.EncodeToPEM()
	if err != nil {
		return nil, err
	}

	// Simple XOR encryption (NOT SECURE - for demo only)
	// TODO: Replace with proper AES-GCM encryption
	encrypted := make([]byte, len(pemData))
	keyIndex := 0
	for i, b := range []byte(pemData) {
		encrypted[i] = b ^ f.encryptionKey[keyIndex%len(f.encryptionKey)]
		keyIndex++
	}

	return encrypted, nil
}

// decryptDPoPKey decrypts a DPoP private key using the configured encryption key.
func (f *FileStorage) decryptDPoPKey(encrypted []byte) (*ecdsa.PrivateKey, error) {
	// Simple XOR decryption (matches encryptDPoPKey)
	// TODO: Replace with proper AES-GCM decryption
	decrypted := make([]byte, len(encrypted))
	keyIndex := 0
	for i, b := range encrypted {
		decrypted[i] = b ^ f.encryptionKey[keyIndex%len(f.encryptionKey)]
		keyIndex++
	}

	// Decode PEM data
	keyPair, err := oauth.DecodeFromPEM(string(decrypted))
	if err != nil {
		return nil, err
	}

	return keyPair.PrivateKey, nil
}

// GetSessionCount returns the number of session files (for testing/debugging).
func (f *FileStorage) GetSessionCount() (int, error) {
	if _, err := os.Stat(f.baseDir); os.IsNotExist(err) {
		return 0, nil
	}

	count := 0
	err := filepath.WalkDir(f.baseDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(d.Name(), ".json") {
			count++
		}
		return nil
	})

	return count, err
}