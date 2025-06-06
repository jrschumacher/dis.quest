package cmd

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"os"

	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/spf13/cobra"
)

var utilCmd = &cobra.Command{
	Use:     "util",
	Aliases: []string{"utils"},
	Short:   "Utility commands for dis.quest",
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Println("Available utility commands:")
		fmt.Println("  generate-jwk - Generate JWKs for the application")
	},
}

var utilGenerateJWKCmd = &cobra.Command{
	Use:   "generate-jwk",
	Short: "Generate JWKs for the application",
	Run: func(_ *cobra.Command, _ []string) {
		// Generate EC P-256 key using crypto/ecdsa
		privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			panic(fmt.Errorf("failed to generate key: %w", err))
		}

		// Wrap as JWK
		key, err := jwk.FromRaw(privKey)
		if err != nil {
			panic(fmt.Errorf("failed to create JWK: %w", err))
		}

		_ = key.Set(jwk.KeyIDKey, "my-app-key")
		_ = key.Set(jwk.AlgorithmKey, jwa.ES256)
		_ = key.Set(jwk.KeyUsageKey, "sig")

		// Export public key only
		pubKey, err := key.PublicKey()
		if err != nil {
			panic(fmt.Errorf("failed to get public key: %w", err))
		}
		pubSet := jwk.NewSet()
		_ = pubSet.AddKey(pubKey)
		pubJSON, _ := json.MarshalIndent(pubSet, "", "  ")
		_ = os.WriteFile("jwks.public.json", pubJSON, 0600)

		// Export private key
		privSet := jwk.NewSet()
		_ = privSet.AddKey(key)
		privJSON, _ := json.MarshalIndent(privSet, "", "  ")
		_ = os.WriteFile("jwks.private.json", privJSON, 0600)

		fmt.Println("JWKs written to jwks.public.json and jwks.private.json")
	},
}

func init() {
	rootCmd.AddCommand(utilCmd)
	utilCmd.AddCommand(utilGenerateJWKCmd)
}
