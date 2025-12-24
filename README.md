TensorBox
TensorBox is a minimal container runtime in Go, designed to bridge WSL 2 GPU (CUDA) into isolated environments. It avoids the complexity of Docker by using Linux namespaces and direct mount propagation.

✨ Core Logic
Namespace Isolation: Implements UTS, PID, and Mount namespaces for process-level isolation.

GPU Passthrough: Leverages WSL 2's /dev/dxg and driver redirection.

Runtime Self-Healing: Automates library discovery for AI frameworks.

🛠️ Usage
Setup Rootfs: Use setup_rootfs.sh to extract a base environment.

Launch: sudo go run main.go run /bin/bash

Verify: python3 -c "import torch; print(torch.cuda.is_available())"