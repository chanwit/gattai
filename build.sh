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

DOCKER_VERSION="v1.9.0-rc4" # origin/master

function update_and_patch {
    local PROJECT_DIR=$1
    local PATCH_FILE=$2
    local REVISION=$3

    ( cd $PROJECT_DIR  &&      \
      git reset --hard HEAD && \
      git remote update &&     \
      git reset --hard $REVISION)

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
    hardlink_cygwin gattai.go


    DOCKER_VENDOR="$GOPATH/src/github.com/docker/docker/vendor"
    # MACHINE_VENDOR="$GOPATH/src/github.com/docker/machine/Godeps/_workspace"
    # OLD_GOPATH=$GOPATH
    if [ "$1" == "" ]; then
        # reset docker
        ( cd $GOPATH/src/github.com/docker/docker  \
          && git remote update                     \
          && git fetch --tags                      \
          && git reset --hard $DOCKER_VERSION      )
          # && git reset --hard origin/master       )

        # reset machine
        ( cd $GOPATH/src/github.com/docker/machine \
          && git add .                             \
          && git remote update                     \
          && git fetch --tags                      \
          && git reset --hard origin/master        )

        # restore
        (cd $GOPATH/src/github.com/docker/machine && $GOPATH/bin/godep restore)
    fi
    export GOPATH="$DOCKER_VENDOR;$GOPATH;$VENDOR"
fi

if [ "$1" == "--cache" ]; then
    GO_FILES=`git status --porcelain | grep .go | awk '{ print $2 }'`
    if [ "$GO_FILES" != "" ]; then
        goimports -w $GO_FILES
    fi
    go install -tags experimental \
       github.com/chanwit/gattai/gattai
    exit 0
fi

if [ "$1" == "--gox" ]; then
    rm gattai_*
    gox -osarch="linux/amd64" github.com/chanwit/gattai/gattai
    xz gattai_*
    exit 0
fi

rm gattai_*

go clean
rm -Rf $GOPATH/pkg
rm -Rf $VENDOR/pkg
# $GOPATH/bin/govers -m github.com/docker/docker github.com/chanwit/docker

update_and_patch ../../docker/docker     001.patch  $DOCKER_VERSION
# update_and_patch ../../docker/libcompose 002.patch
update_and_patch ../../docker/machine    003.patch  origin/master

go install -tags experimental github.com/chanwit/gattai/gattai
if [[ $? -ne 0 ]]; then
    echo  "Build failed"
    exit 1
fi

gox -osarch="windows/amd64 darwin/amd64 linux/amd64 linux/arm" \
    github.com/chanwit/gattai/gattai

echo "Built successsfully"