#!/usr/bin/env bash

# TODO: handle job deployments getting stuck - no restarts/reschedules, or a timeout

set -xe

test -f policy.hcl || {
  echo 'must run from within the demo dir'
  exit 1
}

# create a policy for the csi job to allow it to write Nomad variables
nomad acl policy apply -namespace=default -job=csi-madrid madrid policy.hcl

# check out application logs on exit
trap 'nomad logs -job csi-madrid ; nomad logs -stderr -job csi-madrid' EXIT

# run the csi job
nomad run csi-madrid.nomad.hcl

# wait for plugin to become healthy
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
