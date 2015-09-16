#!/bin/bash

if [ "$(uname)" == "Darwin" ]; then
	echo "Darwin"
elif [ "$(expr substr $(uname -s) 1 5)" == "Linux" ]; then
	echo "Linux"
elif [ "$(expr substr $(uname -s) 1 10)" == "MINGW32_NT" ]; then
    echo "MinGW"
elif [ "$(expr substr $(uname -s) 1 9)" == "CYGWIN_NT" ]; then
	DOCKER_VENDOR="$GOPATH/src/github.com/docker/docker/vendor"
	VENDOR=`cygpath -m $PWD`/vendor
	export GOPATH="$VENDOR;$DOCKER_VENDOR;$GOPATH"
fi

go clean
go install github.com/chanwit/gattai