#!/usr/bin/env bats

load helpers

function setup() {
    _setup
    gattai init
    cat > provision.yml << EOF
---
machines:
  test:
    driver: none
    instances: 2
    options:
      url: tcp://1.2.3.4:2376
EOF

    gattai provision test
}

function teardown() {
    _teardown
}

@test "gattai active - set" {
    run gattai active test-1
    [ "$status" -eq 0 ]

    run stat .gattai/.active_host
    [ "$status" -eq 0 ]

    run cat .gattai/.active_host
    [ "${lines[1]}" = "name: test-1" ]
    [ "${lines[2]}" = "DOCKER_HOST: \"tcp://1.2.3.4:2376\"" ]
}

@test "gattai active - unset" {
    run gattai active test-1
    [ "$status" -eq 0 ]

    run gattai active --
    [ "$status" -eq 0 ]

    run stat .gattai/.active_host
    [ "$status" -ne 0 ]
}

@test "gattai active - get" {
    run gattai active test-1
    [ "$status" -eq 0 ]

    run gattai active
    [ "$status" -eq 0 ]
    [ "${lines[0]}" = "test-1" ]
}

@test "gattai active - switch" {
    run gattai active test-1
    [ "$status" -eq 0 ]

    run gattai active
    [ "$status" -eq 0 ]
    [ "${lines[0]}" = "test-1" ]

    run gattai active test-2
    [ "$status" -eq 0 ]

    run gattai active
    [ "$status" -eq 0 ]
    [ "${lines[0]}" = "test-2" ]

    run gattai active --
    [ "$status" -eq 0 ]

    run gattai active
    [ "$status" -ne 0 ]
    [ "${lines[0]}" = "There is no active host." ]
}
