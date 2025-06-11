#!/usr/bin/env bash

set -xe

sudo $(which nomad) agent -dev -config $PWD/demo/agent.hcl > nomad.log 2>&1 &
trap 'sudo pkill nomad' EXIT

export NOMAD_TOKEN=00000000-0000-0000-0000-000000000000
for _ in {1..30}; do
  nomad acl bootstrap - <<< "$NOMAD_TOKEN" && break
  sleep 1
done

cd demo
./setup.sh
./teardown.sh
