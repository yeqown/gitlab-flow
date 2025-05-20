#!/bin/bash

# install gitlab-flow2 with gitlab application(secretKey?)
# replace follow variables with your personal parameters,
# or export them into ENV variables.
# SECRET_KEY=__YOUR_SECRET_KEY__
#
#
# Usage:
#   SECRET_KEY=gitlab_app_id BIN=gitlab-flow ./install.sh
SECRET_KEY="aflowcli"

if [ -z "${SECRET_KEY}" ]; then
  echo "Warning: SECRET_KEY is empty, using default value: aflowcli"
  $SECRET_KEY="aflowcli"
fi

if [ -z "${BIN}" ]; then
  echo "Warning: BIN is empty, using default value: gitlab-flow"
  BIN="gitlab-flow"
fi

# echo build info and start compiling
echo "Start compiling..."
echo "SECRET_KEY=${SECRET_KEY}"
echo "BIN=${BIN}"

go build \
  -o "${BIN}" \
  -ldflags="-X 'github.com/yeqown/gitlab-flow/internal/gitlab-operator.SecretKey=${SECRET_KEY}'" \
  ./cmd/gitlab-flow

# check if compiled successfully, then install it
echo "Compiled successfully, start installing..."

GOBIN=$(go env GOBIN)
if [ -z "${GOBIN}" ]; then
  echo "Empty GOBIN, installing stopped, please check your go env or move the binary to your PATH manually"
fi

echo "Install ${BIN} to ${GOBIN}"
mv "${BIN}" "${GOBIN}"
