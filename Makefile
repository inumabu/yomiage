.PHONY: build-amd64 archive-amd64 build-windows-amd64 ci

build-amd64:
	TARGET_ARCH=amd64 bash ./build/build.sh

archive-amd64: build-amd64
	tar -C build -czf yomiage-keiryou-linux-amd64.tar.gz dist/yomiage-keiryou-amd64

# WSL uses the DAVE-enabled Linux amd64 binary.
build-windows-amd64: build-amd64
	cp build/dist/yomiage-keiryou-amd64 windows/wsl/yomiage-keiryou-amd64
