package main

import (
	"fmt"
	"os"
	"syscall"

	"tensorbox/pkg/container"
	"tensorbox/pkg/hardware"
	"tensorbox/pkg/isolate"
	"tensorbox/pkg/storage"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Usage: sudo ./tensorbox run /bin/bash\n")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "run":
		parent()
	case "child":
		child()
	default:
		fmt.Println("Unknown command")
	}
}

func parent() {

	id, err := container.GenerateID()
	must(err)

	c, err := container.NewContainer(id, "./rootfs_base")
	must(err)

	must(container.CreateSandboxDirs(c))

	must(storage.MountOverlay(c))

	defer container.RemoveSandboxDirs(c)
	defer storage.UmountOverlay(c)

	userCmdArgs := os.Args[2:]
	cmd := isolate.NewParentProcess(c.MergedDir, userCmdArgs)

	if err := cmd.Run(); err != nil {
		fmt.Printf("Container stopped with error: %v\n", err)
	}
}

func child() {
	// 约定：Args 为 [tensorbox, child, mergedDir, userCmd, userArgs...]
	mergedDir := os.Args[2]
	userCmd := os.Args[3]
	userArgs := os.Args[3:]

	must(syscall.Sethostname([]byte("tensorbox-" + os.Args[2][:6])))
	// 将根目录设为私有挂载，防止容器内的挂载操作泄露到宿主机
	// MS_REC 表示递归应用到所有子目录，MS_PRIVATE 表示不传播挂载事件
	err := syscall.Mount("", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, "")
	if err != nil {
		fmt.Printf("Fatal: failed to set propagation to private: %v\n", err)
		os.Exit(1)
	}
	must(isolate.SetupDevNodes(mergedDir))
	must(hardware.PrepareGPU(mergedDir))
	must(hardware.FixLdCache(mergedDir))

	must(isolate.PivotRoot(mergedDir))

	defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	must(syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), ""))

	env := append(os.Environ(), "LD_LIBRARY_PATH=/usr/lib/wsl/lib")

	path, err := execLookPath(userCmd)
	must(err)

	if err := syscall.Exec(path, userArgs, env); err != nil {
		fmt.Printf("Exec failed: %v\n", err)
	}
}

func execLookPath(cmd string) (string, error) {

	return cmd, nil
}

func must(err error) {
	if err != nil {
		fmt.Printf("Fatal Error: %v\n", err)
		os.Exit(1)
	}
}
