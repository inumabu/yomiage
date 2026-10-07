.PHONY: build-amd64 build-arm64 archive-amd64 archive-arm64 build-windows-amd64 ci

build-amd64:
	TARGET_ARCH=amd64 bash ./build/build.sh

build-arm64:
	TARGET_ARCH=arm64 bash ./build/build.sh

archive-amd64: build-amd64
	tar -C build -czf yomiage-keiryou-linux-amd64.tar.gz dist/yomiage-keiryou-amd64

archive-arm64: build-arm64
	tar -C build -czf yomiage-keiryou-linux-arm64.tar.gz dist/yomiage-keiryou-arm64

# WSL uses the Linux amd64/arm64 binaries above.
build-windows-amd64: build-amd64
	cp build/dist/yomiage-keiryou-amd64 windows/wsl/yomiage-keiryou-amd64
