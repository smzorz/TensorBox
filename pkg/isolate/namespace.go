package isolate

import (
	"os"
	"os/exec"
	"syscall"
)

// NewParentProcess 创建一个准备进入隔离环境的父进程命令对象
// command: 用户要运行的命令 (如 /bin/bash)
// args: 命令参数
// mergedDir: 容器的合并根目录
func NewParentProcess(mergedDir string, args []string) *exec.Cmd {

	childArgs := append([]string{"child", mergedDir}, args...)
	cmd := exec.Command("/proc/self/exe", childArgs...)

	// 设置隔离标志位
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | // 隔离主机名
			syscall.CLONE_NEWPID | // 隔离进程树 (容器内 PID 1)
			syscall.CLONE_NEWNS | // 隔离挂载点 (独立文件系统)
			syscall.CLONE_NEWNET | // 隔离网络 (独立网卡)
			syscall.CLONE_NEWIPC, // 隔离进程间通信
	}

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd
}
