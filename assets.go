//go:build !prod
// +build !prod

package main

import (
	"io/fs"
	"disquest/internal/embedded"
)

func GetStaticAssets() fs.FS {
	return embedded.NewOsFs()
}
