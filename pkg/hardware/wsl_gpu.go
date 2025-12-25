package hardware

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

// PrepareGPU 将 WSL 2 的驱动透传进容器的合并视图中
func PrepareGPU(mergedDir string) error {
	// 1. 挂载驱动库目录
	srcLib := "/usr/lib/wsl"
	targetLib := filepath.Join(mergedDir, "usr/lib/wsl")
	os.MkdirAll(targetLib, 0755)

	if err := syscall.Mount(srcLib, targetLib, "", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
		return fmt.Errorf("GPU lib mount failed: %v", err)
	}

	// 2. 挂载设备节点 /dev/dxg
	dxgSrc := "/dev/dxg"
	dxgTarget := filepath.Join(mergedDir, "dev/dxg")
	if _, err := os.Stat(dxgSrc); err == nil {
		os.MkdirAll(filepath.Dir(dxgTarget), 0755)
		os.Create(dxgTarget)
		syscall.Mount(dxgSrc, dxgTarget, "", syscall.MS_BIND, "")
	}
	return nil
}

func FixLdCache(mergedDir string) error {
	libwsl := filepath.Join(mergedDir, "usr/lib/wsl/lib/libcuda.so")
	target := filepath.Join(mergedDir, "usr/lib/libcuda.so.1")

	if _, err := os.Stat(libwsl); err == nil {
		// --- 新增逻辑：如果目标已存在，先删掉它 ---
		if _, err := os.Lstat(target); err == nil {
			os.Remove(target)
		}

		return os.Symlink("/usr/lib/wsl/lib/libcuda.so", target)
	}
	return nil
}
