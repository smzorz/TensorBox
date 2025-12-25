package container

import (
	"fmt"
	"os"
)

func CreateSandboxDirs(c *Container) error {

	if _, err := os.Stat(c.BaseDir); !os.IsNotExist(err) {
		return fmt.Errorf("container directory %s already exists (ID collision?)", c.BaseDir)
	}

	dirs := []string{c.UpperDir, c.WorkDir, c.MergedDir}
	for _, dir := range dirs {

		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %v", dir, err)
		}
	}
	return nil
}

func RemoveSandboxDirs(c *Container) error {

	return os.RemoveAll(c.BaseDir)
}
