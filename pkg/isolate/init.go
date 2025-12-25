package isolate

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

func SetupDevNodes(mergedDir string) error {
	devPath := filepath.Join(mergedDir, "dev")

	if err := os.MkdirAll(devPath, 0755); err != nil {
		return err
	}
	if err := syscall.Mount("tmpfs", devPath, "tmpfs", syscall.MS_NOSUID|syscall.MS_STRICTATIME, "mode=755"); err != nil {
		return fmt.Errorf("mount tmpfs on /dev failed: %v", err)
	}

	ptsPath := filepath.Join(devPath, "pts")
	os.MkdirAll(ptsPath, 0755)
	if err := syscall.Mount("devpts", ptsPath, "devpts", 0, "newinstance,ptmxmode=0666,mode=0620"); err != nil {
		return fmt.Errorf("mount devpts failed: %v", err)
	}

	shmPath := filepath.Join(devPath, "shm")
	os.MkdirAll(shmPath, 0755)
	if err := syscall.Mount("tmpfs", shmPath, "tmpfs", syscall.MS_NOSUID|syscall.MS_NODEV, "mode=1777,size=65536k"); err != nil {
		return fmt.Errorf("mount /dev/shm failed: %v", err)
	}

	devices := []struct {
		name         string
		mode         uint32
		major, minor uint32
	}{
		{"null", 0666, 1, 3},
		{"zero", 0666, 1, 5},
		{"random", 0666, 1, 8},
		{"urandom", 0666, 1, 9},
	}

	for _, d := range devices {
		path := filepath.Join(devPath, d.name)
		devId := (d.minor & 0xff) | ((d.major & 0xfff) << 8)
		if err := syscall.Mknod(path, uint32(d.mode|syscall.S_IFCHR), int(devId)); err != nil {
			return fmt.Errorf("mknod %s failed: %v", d.name, err)
		}
	}

	os.Symlink("pts/ptmx", filepath.Join(devPath, "ptmx"))
	os.Symlink("/proc/self/fd", filepath.Join(devPath, "fd"))
	os.Symlink("/proc/self/fd/0", filepath.Join(devPath, "stdin"))
	os.Symlink("/proc/self/fd/1", filepath.Join(devPath, "stdout"))
	os.Symlink("/proc/self/fd/2", filepath.Join(devPath, "stderr"))

	fmt.Println(">> [Isolate] Device nodes populated (PTY & SHM enabled).")
	return nil
}

// PivotRoot 切换根文件系统
func PivotRoot(newRoot string) error {

	putOld := filepath.Join(newRoot, ".pivot_root")
	if err := os.MkdirAll(putOld, 0700); err != nil {
		return err
	}

	if err := syscall.PivotRoot(newRoot, putOld); err != nil {
		return fmt.Errorf("pivot_root failed: %v", err)
	}

	if err := os.Chdir("/"); err != nil {
		return err
	}

	putOld = "/.pivot_root"
	if err := syscall.Unmount(putOld, syscall.MNT_DETACH); err != nil {
		return fmt.Errorf("unmount old root failed: %v", err)
	}
	return os.RemoveAll(putOld)
}
