#!/usr/bin/env bash

set -x

nomad stop web
while true; do
  nomad volume status my-vol | grep running || break
  sleep 3
done

nomad volume snapshot delete -secret cooler=secret madrid my-snap
nomad volume delete my-vol

nomad stop csi-madrid

nomad var purge csi-madrid/lock
nomad acl policy delete madrid
