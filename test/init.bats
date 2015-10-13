#!/usr/bin/env bats

function setup() {
  mkdir $BATS_TEST_DIRNAME/temp
}

function teardown() {
  rm -rf $BATS_TEST_DIRNAME/temp
  cd $BATS_TEST_DIRNAME
}

@test "gattai init" {
  cd $BATS_TEST_DIRNAME/temp
  run gattai init
  [ "$status" -eq 0 ]
  [ "${lines[0]}" = "Gattai mission repository is initialized." ]

  run stat .gattai
  [ "$status" -eq 0 ]

  run stat provision.yml
  [ "$status" -eq 0 ]

  run stat composition.yml
  [ "$status" -eq 0 ]
}

@test "gattai re-init" {
  cd $BATS_TEST_DIRNAME/temp
  run gattai init
  [ "$status" -eq 0 ]

  run gattai init
  [ "$status" -ne 0 ]
  [ "${lines[0]}" = ".gattai is already existed" ]
}

@test "gattai re-init with existing provision file" {
  cd $BATS_TEST_DIRNAME/temp
  run gattai init
  [ "$status" -eq 0 ]

  rm -rf .gattai
  echo changed > provision.yml

  run gattai init
  [ "$status" -eq 0 ]

  run cat provision.yml
  [ "${lines[0]}" = "changed" ]
}