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

# check out them logs again after it's all done
nomad logs -job csi-madrid | tail -n40
nomad logs -stderr -job csi-madrid | tail -n40

nomad var purge csi-madrid/lock
nomad acl policy delete madrid
