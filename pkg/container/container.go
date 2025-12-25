package container

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
)

type Container struct {
	ID        string
	BaseDir   string
	LowerDir  string
	UpperDir  string
	WorkDir   string
	MergedDir string
}

func GenerateID() (string, error) {
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate random id: %v", err)
	}
	return hex.EncodeToString(b), nil
}

func NewContainer(id string, lowerDir string) (*Container, error) {
	if id == "" {
		return nil, fmt.Errorf("id cannot be empty")
	}

	absLower, err := filepath.Abs(lowerDir)
	if err != nil {
		return nil, err
	}
	info, err := os.Stat(absLower)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("lowerDir %s does not exist", absLower)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("lowerDir %s is not a directory", absLower)
	}

	base := filepath.Join("/var/lib/tensorbox", id)

	return &Container{
		ID:        id,
		BaseDir:   base,
		LowerDir:  absLower,
		UpperDir:  filepath.Join(base, "upper"),
		WorkDir:   filepath.Join(base, "work"),
		MergedDir: filepath.Join(base, "merged"),
	}, nil
}
