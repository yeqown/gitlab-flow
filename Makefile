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