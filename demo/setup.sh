#!/usr/bin/env bash

set -xeuo pipefail

test -f policy.hcl || {
  echo 'must run from within the demo dir'
  exit 1
}

# create a policy for the csi job to allow it to write Nomad variables
nomad acl policy apply -namespace=default -job=csi-madrid madrid policy.hcl

# run the csi job and wiat for it to become healthy
nomad run csi-madrid.nomad.hcl
while true; do
  nomad plugin status -json madrid \
    | jq '.ControllersHealthy + .NodesHealthy' \
    | grep 2 && break
  sleep 2
done

# create a volume and a snapshot
nomad volume create volume.hcl
nomad volume snapshot create -secret cool=secret my-vol my-snap

# run a job that uses the volume
nomad run job.nomad.hcl
