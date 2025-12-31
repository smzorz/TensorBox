package isolate

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

const cgroupRoot = "/sys/fs/cgroup"

type CgroupManager struct {
	Path string // 例如: /sys/fs/cgroup/tensorbox/container_01
}

func NewCgroupManager(id string) *CgroupManager {
	// 为每个容器创建一个唯一的 cgroup 路径
	path := filepath.Join(cgroupRoot, "tensorbox", id)
	return &CgroupManager{Path: path}
}

// Init 准备 cgroup 目录
func (m *CgroupManager) Init() error {
	// 1. 创建父目录 /sys/fs/cgroup/tensorbox
	parentDir := filepath.Dir(m.Path) // 即 /sys/fs/cgroup/tensorbox
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return err
	}

	// 在父级目录开启控制器授权
	// 只有向这里写入 "+memory"，子目录才能使用内存限制功能
	subtreePath := filepath.Join(parentDir, "cgroup.subtree_control")

	_ = os.WriteFile(subtreePath, []byte("+memory +cpu +pids"), 0644)

	return os.MkdirAll(m.Path, 0755)
}

// SetMemoryLimit 设置内存限制，例如 "2G" 或 "512M"

func (m *CgroupManager) SetMemoryLimit(limit string) error {
	if limit == "" {
		return nil
	}
	// 限制物理内存
	memPath := filepath.Join(m.Path, "memory.max")
	if err := os.WriteFile(memPath, []byte(limit), 0644); err != nil {
		return err
	}

	// 【新增】禁用 Swap 交换。设为 0 表示该控制组不允许使用交换分区
	swapPath := filepath.Join(m.Path, "memory.swap.max")
	_ = os.WriteFile(swapPath, []byte("0"), 0644)

	return nil
}

// SetCPULimit 限制 CPU 使用率

func (m *CgroupManager) SetCPULimit(limit float64) error {

	if limit == 0 {
		return nil
	}

	const period = 100000

	quota := int(limit * period)

	limitString := fmt.Sprintf("%d %d", quota, period)

	cpuPath := filepath.Join(m.Path, "cpu.max")
	if err := os.WriteFile(cpuPath, []byte(limitString), 0644); err != nil {
		return fmt.Errorf("failed to set cpu limit: %v", err)
	}

	return nil
}

// Apply 将进程加入该控制组
func (m *CgroupManager) Apply(pid int) error {
	procPath := filepath.Join(m.Path, "cgroup.procs")
	return os.WriteFile(procPath, []byte(strconv.Itoa(pid)), 0644)
}

// Destroy 清理 cgroup 目录
func (m *CgroupManager) Destroy() error {
	// 只有目录为空且没有进程驻留时才能成功删除
	return os.RemoveAll(m.Path)
}
