#!/bin/bash

if [ "$(uname)" == "Darwin" ]; then
	echo "Darwin"
elif [ "$(expr substr $(uname -s) 1 5)" == "Linux" ]; then
	echo "Linux"
elif [ "$(expr substr $(uname -s) 1 10)" == "MINGW32_NT" ]; then
    echo "MinGW"
elif [ "$(expr substr $(uname -s) 1 9)" == "CYGWIN_NT" ]; then
	VENDOR=`cygpath -m $PWD`/vendor
	PACKAGE_PATH="github.com/docker/docker/api/client"

	fsutil hardlink list $GOPATH/src/$PACKAGE_PATH/provision.go > /dev/null
	if [[ $? -ne 0 ]]; then
	fsutil hardlink create \
		$GOPATH/src/$PACKAGE_PATH/provision.go \
		$VENDOR/src/$PACKAGE_PATH/provision.go
	fi

	if [[ $? -ne 0 ]]; then
	fsutil hardlink list $GOPATH/src/$PACKAGE_PATH/up.go > /dev/null
	fsutil hardlink create \
		$GOPATH/src/$PACKAGE_PATH/up.go \
		$VENDOR/src/$PACKAGE_PATH/up.go
	fi

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

(cd ../../docker/docker     && git remote update && git reset --hard origin/master)
(cd ../../docker/libcompose && git remote update && git reset --hard origin/master)

patch --dry-run -p1 -d ../../docker/docker -f < 001.patch
if [[ $? -ne 0 ]]; then
	echo "Patch not successsfully, aborted"
	exit 1
fi

patch --dry-run -p1 -d ../../docker/libcompose -f < 002.patch
if [[ $? -ne 0 ]]; then
	echo "Patch not successsfully, aborted"
	exit 1
fi

patch -p1 -d ../../docker/docker -f < 001.patch
patch -p1 -d ../../docker/libcompose -f < 002.patch
go install github.com/chanwit/gattai
echo "Built successsfully"