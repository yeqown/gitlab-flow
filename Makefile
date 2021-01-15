build:
	go build -o ${BIN} ./cmd/gitlab-flow

debug:
	BIN=flow-debug make build
	cp ./flow-debug ~/go/bin

pre-release:
	BIN=flow2 make build
	cp ./flow2 ~/go/bin