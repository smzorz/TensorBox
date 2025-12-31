package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"tensorbox/pkg/container"
	"tensorbox/pkg/hardware"
	"tensorbox/pkg/isolate"
	"tensorbox/pkg/storage"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: tensorbox run -m [limit] [command]")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "run":

		runCmd := flag.NewFlagSet("run", flag.ExitOnError)

		memLimit := runCmd.String("m", "512m", "Memory limit")

		cpuLimit := runCmd.Float64("cpu", 0, "CPU limit (e.g., 0.5, 1.0)")

		runCmd.Parse(os.Args[2:])

		// 传给 parent
		parent(*memLimit, *cpuLimit, runCmd.Args())

	case "child":
		child() // child 逻辑保持不变
	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}

func parent(memLimit string, cpuLimit float64, args []string) {

	if len(args) < 1 {
		fmt.Println("Error: No command specified. Usage: tensorbox run [command]")
		return
	}

	id, err := container.GenerateID()
	must(err)

	cg := isolate.NewCgroupManager(id)
	must(cg.Init())
	defer cg.Destroy()

	fmt.Printf(">> [Cgroup] Setting memory limit to: %s\n", memLimit)
	must(cg.SetMemoryLimit(memLimit))

	if cpuLimit > 0 {
		fmt.Printf(">> [Cgroup] Setting CPU limit to: %.1f core(s)\n", cpuLimit)
		must(cg.SetCPULimit(cpuLimit))
	}

	c, err := container.NewContainer(id, "./rootfs_base")
	must(err)

	must(container.CreateSandboxDirs(c))
	must(storage.MountOverlay(c))

	defer container.RemoveSandboxDirs(c)
	defer storage.UmountOverlay(c)

	cmd := isolate.NewParentProcess(c.MergedDir, args)

	if err := cmd.Start(); err != nil {
		fmt.Printf("Failed to start: %v\n", err)
		return
	}

	fmt.Printf(">> [Cgroup] Target PID: %d\n", cmd.Process.Pid)

	fmt.Printf(">> [Cgroup] Limiting container (PID: %d) to %s\n", cmd.Process.Pid, memLimit)
	must(cg.Apply(cmd.Process.Pid))

	if err := cmd.Wait(); err != nil {
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

	if len(cmd) > 0 && cmd[0] == '/' {
		return cmd, nil
	}

	return exec.LookPath(cmd)
}

func must(err error) {
	if err != nil {
		fmt.Printf("Fatal Error: %v\n", err)
		os.Exit(1)
	}
}
