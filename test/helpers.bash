#!/bin/bash

function _setup() {
  mkdir $BATS_TEST_DIRNAME/temp
  cd $BATS_TEST_DIRNAME/temp
}

function _teardown() {
  rm -rf $BATS_TEST_DIRNAME/temp
  cd $BATS_TEST_DIRNAME
}
