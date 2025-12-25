#!/bin/bash
# TensorBox Rootfs 构建脚本

set -e 

ROOTFS_DIR="rootfs"

echo ">> [1/4] Cleaning up old rootfs..."
sudo rm -rf $ROOTFS_DIR
mkdir -p $ROOTFS_DIR

echo ">> [2/4] Extracting official Ubuntu image using Docker..."

docker pull ubuntu:22.04
TEMP_CONTAINER=$(docker create ubuntu:22.04)
docker export $TEMP_CONTAINER | tar -C $ROOTFS_DIR -x
docker rm $TEMP_CONTAINER

echo ">> [3/4] Injecting Host DNS configuration..."

sudo cp /etc/resolv.conf $ROOTFS_DIR/etc/resolv.conf

echo ">> [4/4] Initializing GPU driver directories..."

sudo mkdir -p $ROOTFS_DIR/usr/lib/wsl
sudo mkdir -p $ROOTFS_DIR/dev/dxg

echo "------------------------------------------------"
echo "Rootfs preparation complete!"
echo "Now you can run: sudo go run main.go run /bin/bash"
echo "Inside the container, run 'apt-get update' to verify."