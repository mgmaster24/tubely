package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"strings"
)

func (cfg apiConfig) ensureAssetsDir() error {
	if _, err := os.Stat(cfg.assetsRoot); os.IsNotExist(err) {
		return os.Mkdir(cfg.assetsRoot, 0755)
	}
	return nil
}

func getAssetsPath(mediaType string) string {
	randBytes := make([]byte, 32)
	_, err := rand.Read(randBytes)
	if err != nil {
		panic("failed to generate random bytes")
	}

	id := base64.RawURLEncoding.EncodeToString(randBytes)
	return fmt.Sprintf("%s%s", id, getExtFromMediaType(mediaType))
}

func getExtFromMediaType(mediaType string) string {
	parts := strings.Split(mediaType, "/")
	if len(parts) != 2 {
		return ".bin"
	}

	return "." + parts[1]
}
