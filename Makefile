GOBIN=`go env GOBIN`

build:
	# install gitlab-flow2 with gitlab application(appId, appSecret)
	# replace follow variables with your personal parameters,
	# or export them into ENV variables.
	# APP_ID=__YOUR_APP_ID__
	# APP_SECRET=__YOUR_APP_SECRET__

	@ APP_ID=$APP_ID
	@ APP_SECRET=$APP_SECRET
	@ BIN=$BIN

	@ if [ -z ${APP_ID} ]; then \
	  echo "Empty APP_ID, !!!replace your APP_ID at first"; \
	  exit 1; \
	fi

	@ if [ -z ${APP_SECRET} ]; then \
	  echo "Empty APP_SECRET, !!!replace your APP_SECRET at first"; \
	  exit 1; \
	fi

	@ if [ -z ${BIN} ]; then \
   	  echo "Empty BIN, use default name: gitlab-flow"; \
	  $BIN="gitlab-flow";\
	fi

	@ echo "APP_ID=${APP_ID}"
	@ echo "APP_SECRET=${APP_SECRET}"
	@ echo "BIN=${BIN}"
	@ echo "Start building..."

	go build \
		-o ${BIN} \
		-ldflags="-X 'github.com/yeqown/gitlab-flow/internal/gitlab-operator.OAuth2AppID=${APP_ID}' \
				  -X 'github.com/yeqown/gitlab-flow/internal/gitlab-operator.OAuth2AppSecret=${APP_SECRET}'" \
		./cmd/gitlab-flow

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