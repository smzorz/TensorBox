package storage

import (
	"fmt"
	"path/filepath"
	"syscall"
	"tensorbox/pkg/container"
)

func MountOverlay(c *container.Container) error {
	// 拼接参数：lowerdir=...,upperdir=...,workdir=...
	opts := fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s",
		c.LowerDir, c.UpperDir, c.WorkDir)

	// syscall.Mount(source, target, fstype, flags, data)

	err := syscall.Mount("overlay", c.MergedDir, "overlay", 0, opts)
	if err != nil {
		return fmt.Errorf("overlay mount failed: %v", err)
	}
	return nil
}

// UmountOverlay 卸载合并目录
func UmountOverlay(c *container.Container) error {
	// Try to unmount common nested mounts created inside the container root.
	// Ignore errors for mounts that don't exist or are already gone.
	subs := []string{
		filepath.Join(c.MergedDir, "proc"),
		filepath.Join(c.MergedDir, "dev", "pts"),
		filepath.Join(c.MergedDir, "dev", "shm"),
		filepath.Join(c.MergedDir, "dev"),
		filepath.Join(c.MergedDir, "usr", "lib", "wsl"), // 建议加上这一行
	}
	for _, s := range subs {
		_ = syscall.Unmount(s, syscall.MNT_DETACH)
	}

	if err := syscall.Unmount(c.MergedDir, syscall.MNT_DETACH); err != nil {
		return fmt.Errorf("overlay umount failed: %v", err)
	}
	return nil
}
