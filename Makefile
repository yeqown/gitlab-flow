changelog:
	# `npm install -g conventional-changelog-cli` to install the CLI.
	# changelog all
	# conventional-changelog -p angular -i CHANGELOG.md -s -r 0

	# only newer changes
	conventional-changelog -p angular -i CHANGELOG.md -s

debug:
	BIN=flow-debug make build
	mv ./flow-debug ${GOBIN}

pre-release:
	echo "GOBIN: ${GOBIN}"
	# install.local.sh export APP_ID and APP_SECRET
	BIN=flow2 bash ./install.local.sh
	mv ./flow2 ${GOBIN}

test-build-all:
	# 测试跨平台编译 linux darwin windows 的 amd64, arm64 版本
	mkdir -p ./build
	# 编译 linux 版本
	GOOS=linux GOARCH=amd64 go build -o ./build/flow2-linux-amd64 ./cmd/gitlab-flow || true
	GOOS=linux GOARCH=arm64 go build -o ./build/flow2-linux-arm64 ./cmd/gitlab-flow || true

	# 编译 darwin 版本
	GOOS=darwin GOARCH=amd64 go build -o ./build/flow2-darwin-amd64 ./cmd/gitlab-flow || true
	GOOS=darwin GOARCH=arm64 go build -o ./build/flow2-darwin-arm64 ./cmd/gitlab-flow || true

	# 编译 windows 版本
	GOOS=windows GOARCH=amd64 go build -o ./build/flow2-windows-amd64 ./cmd/gitlab-flow || true
	GOOS=windows GOARCH=arm64 go build -o ./build/flow2-windows-arm64 ./cmd/gitlab-flow || true