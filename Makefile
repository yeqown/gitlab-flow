build:
	go build -o ${BIN} ./cmd/gitlab-flow

debug:
	BIN=flow-debug make build
	cp ./flow-debug ~/go/bin