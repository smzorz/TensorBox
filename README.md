# TensorBox

TensorBox is a minimal, daemon-less container runtime written in Go. It is designed to run GPU-accelerated workloads (for example, PyTorch) inside an isolated environment on WSL 2 while keeping the runtime lightweight and dependency-free.

## 🚀 Features

- Single binary: no background daemon required.
- WSL2 GPU passthrough: projects WSL2 NVIDIA driver libraries and the `/dev/dxg` device into the container.
- Small & focused: uses Linux namespaces, bind mounts and cgroups to isolate processes.

## 🛠️ Prerequisites

- Host: WSL 2 (Ubuntu distribution recommended) with NVIDIA WSL-compatible drivers installed.
- A host NVIDIA GPU with working `libcuda` / driver (verify with `nvidia-smi`).
- `sudo` privileges to run namespace and mount operations.
- If building from source: Go 1.20+ installed on the host.

## 1. Prepare rootfs

TensorBox runs a minimal root filesystem under `rootfs/`. Use the provided script to prepare it (the script uses Docker to export a clean Ubuntu tarball):

```bash
chmod +x setup_rootfs.sh
./setup_rootfs.sh
```

After the script finishes you should have a usable `rootfs/` directory.

## 2. Build

Build the runtime binary (optional: you can also run with `go run`):

```bash
go build -o tensorbox main.go
```

## 3. Run

Start an interactive shell inside a TensorBox container:

```bash
sudo ./tensorbox run /bin/bash
```

Inside the container you will be dropped into the chroot environment from `rootfs/`.

## 4. Verify GPU support

Run these checks inside the container to confirm GPU access:

```bash
# device node
ls -l /dev/dxg /dev/nvidia* 2>/dev/null || true

# check driver library
ldconfig -p | grep -i libcuda || true

# python / PyTorch
python3 - <<'PY'
import torch
print('torch:', torch.__version__)
print('cuda available:', torch.cuda.is_available())
print('cuda version:', torch.version.cuda)
print('device count:', torch.cuda.device_count())
PY
```

If `cuda available` is `True`, PyTorch can access the GPU.


## How it works (brief)

- The runtime creates new UTS/PID/mount namespaces, sets up device nodes under `rootfs/dev`, bind-mounts the WSL driver libraries and `/dev/dxg`, performs a `chroot` and launches the requested process.
- It also attempts to ensure `libcuda` is discoverable by the dynamic loader.

## Contributing

- Issues and PRs welcome. Please open issues for bugs or feature requests.

## License

MIT