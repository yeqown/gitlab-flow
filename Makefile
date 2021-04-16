build:
	go build -o ${BIN} ./cmd/gitlab-flow

changelog:
	# `npm install -g conventional-changelog-cli` to install the CLI.
	# changelog all
	# conventional-changelog -p angular -i CHANGELOG.md -s -r 0

	# only newer changes
	conventional-changelog -p angular -i CHANGELOG.md -s

debug:
	BIN=flow-debug make build
	mv ./flow-debug ~/go/bin

pre-release:
	BIN=flow2 make build
	mv ./flow2 ~/go/bin