build:
	go build -o ${BIN} ./cmd/gitlab-flow

debug:
	BIN=flow-debug make build
	mv ./flow-debug ~/go/bin

pre-release:
	BIN=flow2 make build
	mv ./flow2 ~/go/bin