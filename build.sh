#!/bin/bash

if [ "$(uname)" == "Darwin" ]; then
	echo "Darwin"
elif [ "$(expr substr $(uname -s) 1 5)" == "Linux" ]; then
	echo "Linux"
elif [ "$(expr substr $(uname -s) 1 10)" == "MINGW32_NT" ]; then
    echo "MinGW"
elif [ "$(expr substr $(uname -s) 1 9)" == "CYGWIN_NT" ]; then
	DOCKER_VENDOR="$GOPATH/src/github.com/docker/docker/vendor"
	# MACHINE_VENDOR="$GOPATH/src/github.com/docker/machine/Godeps/_workspace"
	VENDOR=`cygpath -m $PWD`/vendor
	# OLD_GOPATH=$GOPATH
	(cd $GOPATH/src/github.com/docker/machine && $GOPATH/bin/godep restore)
	export GOPATH="$DOCKER_VENDOR;$GOPATH;$VENDOR"
fi

go clean
rm -Rf $GOPATH/pkg
rm -Rf $VENDOR/pkg
# $GOPATH/bin/govers -m github.com/docker/docker github.com/chanwit/docker

(cd ../../docker/docker && git remote update && git reset --hard HEAD)
patch --dry-run -p1 -d ../../docker/docker -f < 001.patch
if [ "$?" == "0" ]; then
	patch -p1 -d ../../docker/docker -f < 001.patch
	go install github.com/chanwit/gattai
else
	echo "patch not successsfully, aborted"
fi