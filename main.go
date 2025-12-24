package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Usage: sudo go run main.go run /bin/bash\n")
		os.Exit(1)
	}
	switch os.Args[1] {
	case "run":
		run()
	case "child":
		child()
	}
}

func run() {
	cmd := exec.Command("/proc/self/exe", append([]string{"child"}, os.Args[2:]...)...)
	cmd.SysProcAttr = &syscall.SysProcAttr{

		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
	}
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	must(cmd.Run())
}

func child() {
	must(syscall.Sethostname([]byte("tensorbox")))
	setupCgroups()

	pwd, _ := os.Getwd()
	rootfs := filepath.Join(pwd, "rootfs")

	setupDevNodes(rootfs)

	mountGPU(rootfs)

	must(syscall.Chroot(rootfs))
	must(os.Chdir("/"))

	must(syscall.Mount("proc", "proc", "proc", 0, ""))

	fixSymlinks()

	cmd := exec.Command(os.Args[2], os.Args[3:]...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr

	cmd.Env = append(os.Environ(), "LD_LIBRARY_PATH=/usr/lib/wsl/lib")

	must(cmd.Run())

	syscall.Unmount("/proc", 0)
}

func mountGPU(rootfs string) {
	fmt.Println(">> [GPU] Binding WSL Driver & DXG (Recursive)...")

	srcLib := "/usr/lib/wsl"
	targetLib := filepath.Join(rootfs, "usr/lib/wsl")
	os.MkdirAll(targetLib, 0755)

	if err := syscall.Mount(srcLib, targetLib, "", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
		fmt.Printf("Warning: GPU lib mount failed: %v\n", err)
	}

	dxgSrc := "/dev/dxg"
	dxgTarget := filepath.Join(rootfs, "dev/dxg")
	if _, err := os.Stat(dxgSrc); err == nil {
		os.MkdirAll(filepath.Dir(dxgTarget), 0755)
		os.Create(dxgTarget)

		syscall.Mount(dxgSrc, dxgTarget, "", syscall.MS_BIND, "")
	}
}

func setupDevNodes(rootfs string) {

	fmt.Println(">> [Dev] Populating device nodes...")
	devPath := filepath.Join(rootfs, "dev")
	os.MkdirAll(devPath, 0755)
	syscall.Mount("tmpfs", devPath, "tmpfs", syscall.MS_NOSUID|syscall.MS_STRICTATIME, "mode=755")

	devices := []struct {
		name string
		m, n int
	}{
		{"null", 1, 3},
		{"zero", 1, 5},
		{"random", 1, 8},
		{"urandom", 1, 9},
	}
	for _, d := range devices {
		path := filepath.Join(devPath, d.name)
		dev := int((d.m << 8) | (d.n & 0xff))
		syscall.Mknod(path, uint32(0666|syscall.S_IFCHR), dev)
		os.Chmod(path, 0666)
	}

	fmt.Println(">> [Dev] Mounting /tmp as tmpfs...")
	tmpPath := filepath.Join(rootfs, "tmp")

	os.MkdirAll(tmpPath, 0755)

	must(syscall.Mount("tmpfs", tmpPath, "tmpfs", 0, "mode=1777"))
}

func fixSymlinks() {

	libwsl := "/usr/lib/wsl/lib/libcuda.so"
	target := "/usr/lib/libcuda.so.1"
	if _, err := os.Stat(libwsl); err == nil {
		os.Symlink(libwsl, target)
		os.Symlink(libwsl, "/usr/lib/libcuda.so")
	}
}

func setupCgroups() {
	cgroupPath := "/sys/fs/cgroup/pids/tensorbox"
	os.MkdirAll(cgroupPath, 0755)
	os.WriteFile(filepath.Join(cgroupPath, "pids.max"), []byte("50"), 0700)
	os.WriteFile(filepath.Join(cgroupPath, "cgroup.procs"), []byte(strconv.Itoa(os.Getpid())), 0700)
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
