#!/usr/bin/env bats

load helpers

function setup() {
    _setup
    gattai init
    cat > provision.yml << EOF
---
machines:
EOF
}

function teardown() {
    _teardown
}

@test "gattai add -f not-existed" {
    run gattai add --flavor not-existed test
    [ "$status" -ne 0 ]
}

@test "gattai add --flavor none" {
    run gattai add --flavor none test
    [ "$status" -eq 0 ]

    run gattai provision test
    echo $output
    [ "$status" -eq 0 ]

    run gattai active test
    [ "$status" -eq 0 ]

    run stat .gattai/.active_host
    [ "$status" -eq 0 ]

    run cat .gattai/.active_host
    [ "${lines[1]}" = "name: test" ]
    [ "${lines[2]}" = "DOCKER_HOST: \"tcp://1.2.3.4:2376\"" ]
}

@test "gattai add --flavor none with --instances 2" {
    run gattai add --flavor none --instances 2 test
    echo $output
    [ "$status" -eq 0 ]

    run gattai provision test
    echo $output
    [ "$status" -eq 0 ]

    run gattai active test-1
    [ "$status" -eq 0 ]

    run stat .gattai/.active_host
    [ "$status" -eq 0 ]

    run cat .gattai/.active_host
    [ "${lines[1]}" = "name: test-1" ]
    [ "${lines[2]}" = "DOCKER_HOST: \"tcp://1.2.3.4:2376\"" ]

    run gattai active test-2
    [ "$status" -eq 0 ]

    run stat .gattai/.active_host
    [ "$status" -eq 0 ]

    run cat .gattai/.active_host
    [ "${lines[1]}" = "name: test-2" ]
    [ "${lines[2]}" = "DOCKER_HOST: \"tcp://1.2.3.4:2376\"" ]
}

@test "gattai add do-2g" {
    run gattai add --flavor do-2g test
    [ "$status" -eq 0 ]

    run cat provision.yml
    [ "${lines[1]}" = "  test:" ]
    [ "${lines[2]}" = "    driver: digitalocean" ]
    [ "${lines[3]}" = "    instances: 1" ]
    [ "${lines[4]}" = "    options:" ]
    [ "${lines[5]}" = "      digitalocean-access-token: \$DIGITALOCEAN_ACCESS_TOKEN" ]
    [ "${lines[6]}" = "      digitalocean-image: ubuntu-14-04-x64" ]
    [ "${lines[7]}" = "      digitalocean-region: nyc3" ]
    [ "${lines[8]}" = "      digitalocean-size: 2gb" ]
    [ "${lines[9]}" = "      engine-install-url: https://get.docker.com" ]
}

@test "gattai add digitalocean-2g" {
    run gattai add --flavor digitalocean-2g test
    [ "$status" -eq 0 ]

    run cat provision.yml
    [ "${lines[1]}" = "  test:" ]
    [ "${lines[2]}" = "    driver: digitalocean" ]
    [ "${lines[3]}" = "    instances: 1" ]
    [ "${lines[4]}" = "    options:" ]
    [ "${lines[5]}" = "      digitalocean-access-token: \$DIGITALOCEAN_ACCESS_TOKEN" ]
    [ "${lines[6]}" = "      digitalocean-image: ubuntu-14-04-x64" ]
    [ "${lines[7]}" = "      digitalocean-region: nyc3" ]
    [ "${lines[8]}" = "      digitalocean-size: 2gb" ]
    [ "${lines[9]}" = "      engine-install-url: https://get.docker.com" ]
}

@test "gattai add do-2g-exp" {
    run gattai add --flavor do-2g-exp test
    [ "$status" -eq 0 ]

    run cat provision.yml
    [ "${lines[1]}" = "  test:" ]
    [ "${lines[2]}" = "    driver: digitalocean" ]
    [ "${lines[3]}" = "    instances: 1" ]
    [ "${lines[4]}" = "    options:" ]
    [ "${lines[5]}" = "      digitalocean-access-token: \$DIGITALOCEAN_ACCESS_TOKEN" ]
    [ "${lines[6]}" = "      digitalocean-image: debian-8-x64" ]
    [ "${lines[7]}" = "      digitalocean-region: nyc3" ]
    [ "${lines[8]}" = "      digitalocean-size: 2gb" ]
    [ "${lines[9]}" = "      engine-install-url: https://experimental.docker.com" ]
}

@test "gattai add digitalocean-2g-exp" {
    run gattai add --flavor digitalocean-2g-exp test
    [ "$status" -eq 0 ]

    run cat provision.yml
    [ "${lines[1]}" = "  test:" ]
    [ "${lines[2]}" = "    driver: digitalocean" ]
    [ "${lines[3]}" = "    instances: 1" ]
    [ "${lines[4]}" = "    options:" ]
    [ "${lines[5]}" = "      digitalocean-access-token: \$DIGITALOCEAN_ACCESS_TOKEN" ]
    [ "${lines[6]}" = "      digitalocean-image: debian-8-x64" ]
    [ "${lines[7]}" = "      digitalocean-region: nyc3" ]
    [ "${lines[8]}" = "      digitalocean-size: 2gb" ]
    [ "${lines[9]}" = "      engine-install-url: https://experimental.docker.com" ]
}

@test "gattai add do-2g-cluster" {
    run gattai add --flavor do-2g-cluster test

    run wc -l provision.yml
    [ "$output" = "25 provision.yml" ]

    run cat provision.yml

    [ "${lines[1]}"  = "  test:" ]
    [ "${lines[2]}"  = "    driver: digitalocean" ]
    [ "${lines[3]}"  = "    instances: 1" ]
    [ "${lines[4]}"  = "    options:" ]
    [ "${lines[5]}"  = "      digitalocean-access-token: \$DIGITALOCEAN_ACCESS_TOKEN" ]
    [ "${lines[6]}"  = "      digitalocean-image: debian-8-x64" ]
    [ "${lines[7]}"  = "      digitalocean-region: nyc3" ]
    [ "${lines[8]}"  = "      digitalocean-size: 2gb" ]
    [ "${lines[9]}"  = "      engine-install-url: https://experimental.docker.com" ]
    [ "${lines[10]}" = "    cluster-store: test-master" ]
    [ "${lines[11]}" = "    post-provision:" ]
    [ "${lines[12]}" = "    - docker network create -d overlay multihost" ]

    [ "${lines[13]}" = "  test-master:" ]
    [ "${lines[14]}" = "    driver: digitalocean" ]
    [ "${lines[15]}" = "    instances: 1" ]
    [ "${lines[16]}" = "    options:" ]
    [ "${lines[17]}" = "      digitalocean-access-token: \$DIGITALOCEAN_ACCESS_TOKEN" ]
    [ "${lines[18]}" = "      digitalocean-image: debian-8-x64" ]
    [ "${lines[19]}" = "      digitalocean-region: nyc3" ]
    [ "${lines[20]}" = "      digitalocean-size: 2gb" ]
    [ "${lines[21]}" = "      engine-install-url: https://experimental.docker.com" ]
    [ "${lines[22]}" = "    post-provision:" ]
    [ "${lines[23]}" = "    - docker run -d -p 8400:8400 -p 8500:8500 -p 8600:53/udp progrium/consul --server" ]
    [ "${lines[24]}" = "      -bootstrap-expect 1" ]

}