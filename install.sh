#!/bin/bash

# install gitlab-flow2 with gitlab application(appId, appSecret)
# replace follow variables with your personal parameters,
# or export them into ENV variables.
# APP_ID=__YOUR_APP_ID__
# APP_SECRET=__YOUR_APP_SECRET__
#
#
# Usage:
#   APP_ID=gitlab_app_id APP_SECRET=gitlab_app_secret BIN=gitlab-flow ./install.sh

APP_ID=$APP_ID
APP_SECRET=$APP_SECRET
BIN=$BIN

if [ -z ${APP_ID} ]; then
  echo "Empty APP_ID, !!!replace your APP_ID at first"
  exit 1
fi

if [ -z ${APP_SECRET} ]; then
  echo "Empty APP_SECRET, !!!replace your APP_SECRET at first"
  exit 1
fi

if [ -z ${BIN} ]; then
  echo "Empty BIN, using default name: gitlab-flow"
  BIN="gitlab-flow"
fi

# echo build info and start compiling
echo "Start compiling..."
echo "APP_ID=${APP_ID}"
echo "APP_SECRET=${APP_SECRET}"
echo "BIN=${BIN}"

go build \
  -o ${BIN} \
  -ldflags="-X 'github.com/yeqown/gitlab-flow/internal/gitlab-operator.OAuth2AppID=${APP_ID}' \
            -X 'github.com/yeqown/gitlab-flow/internal/gitlab-operator.OAuth2AppSecret=${APP_SECRET}'" \
  ./cmd/gitlab-flow

# check if compiled successfully, then install it
echo "Compiled successfully, start installing..."

GOBIN=`go env GOBIN`
if [ -z ${GOBIN} ]; then
  echo "Empty GOBIN, installing stopped, please check your go env or move the binary to your PATH manually"
fi

echo "Install ${BIN} to ${GOBIN}"
mv ${BIN} ${GOBIN}
