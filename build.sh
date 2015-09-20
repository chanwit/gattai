#!/bin/bash

PACKAGE_PATH="github.com/docker/docker/api/client"

function hardlink_cygwin {
	for file in "$@"
	do
		fsutil hardlink list \
			$GOPATH/src/$PACKAGE_PATH/$file > /dev/null
		if [[ $? -ne 0 ]]; then
			fsutil hardlink create \
				$GOPATH/src/$PACKAGE_PATH/$file \
				$VENDOR/src/$PACKAGE_PATH/$file
		fi
	done
}

function update_and_patch {
	local PROJECT_DIR=$1
	local PATCH_FILE=$2

	(cd $PROJECT_DIR  && git remote update && git reset --hard origin/master)

	patch --dry-run -p1 -d $PROJECT_DIR -f < $PATCH_FILE
	if [[ $? -eq 0 ]]; then
		patch -p1 -d $PROJECT_DIR -f < $PATCH_FILE
	else
		echo "Patch not successsfully, aborted"
	fi
}

if [ "$(uname)" == "Darwin" ]; then
	echo "Darwin"
elif [ "$(expr substr $(uname -s) 1 5)" == "Linux" ]; then
	echo "Linux"
elif [ "$(expr substr $(uname -s) 1 10)" == "MINGW32_NT" ]; then
    echo "MinGW"
elif [ "$(expr substr $(uname -s) 1 9)" == "CYGWIN_NT" ]; then
	VENDOR=`cygpath -m $PWD`/vendor
	hardlink_cygwin  \
		active.go    \
		init.go      \
		provision.go \
		up.go

	DOCKER_VENDOR="$GOPATH/src/github.com/docker/docker/vendor"
	# MACHINE_VENDOR="$GOPATH/src/github.com/docker/machine/Godeps/_workspace"
	# OLD_GOPATH=$GOPATH
	if [ "$1" == "" ]; then
		(cd $GOPATH/src/github.com/docker/machine && $GOPATH/bin/godep restore)
	fi
	export GOPATH="$DOCKER_VENDOR;$GOPATH;$VENDOR"
fi

if [ "$1" == "--cache" ]; then
	go install github.com/chanwit/gattai
	exit 0
fi

go clean
rm -Rf $GOPATH/pkg
rm -Rf $VENDOR/pkg
# $GOPATH/bin/govers -m github.com/docker/docker github.com/chanwit/docker

update_and_patch ../../docker/docker     001.patch
update_and_patch ../../docker/libcompose 002.patch

go install github.com/chanwit/gattai
echo "Built successsfully"