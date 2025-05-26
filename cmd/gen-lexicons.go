package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var genLexiconsCmd = &cobra.Command{
	Use:   "gen-lexicons",
	Short: "Generate pretty-printed lexicon JSON files under api/disquest",
	Run: func(cmd *cobra.Command, args []string) {
		lexiconDir := "lexicons"
		outputDir := "api/disquest"

		if err := os.MkdirAll(outputDir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create output dir: %v\n", err)
			os.Exit(1)
		}

		files, err := ioutil.ReadDir(lexiconDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to read lexicon dir: %v\n", err)
			os.Exit(1)
		}

		for _, file := range files {
			if file.IsDir() || filepath.Ext(file.Name()) != ".json" {
				continue
			}
			inPath := filepath.Join(lexiconDir, file.Name())
			outPath := filepath.Join(outputDir, file.Name())

			data, err := ioutil.ReadFile(inPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to read %s: %v\n", inPath, err)
				continue
			}

			var v interface{}
			if err := json.Unmarshal(data, &v); err != nil {
				fmt.Fprintf(os.Stderr, "Invalid JSON in %s: %v\n", inPath, err)
				continue
			}

			pretty, err := json.MarshalIndent(v, "", "  ")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to marshal %s: %v\n", inPath, err)
				continue
			}

			if err := ioutil.WriteFile(outPath, pretty, 0644); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to write %s: %v\n", outPath, err)
				continue
			}
			fmt.Printf("Wrote %s\n", outPath)
		}
	},
}

func init() {
	rootCmd.AddCommand(genLexiconsCmd)
}
