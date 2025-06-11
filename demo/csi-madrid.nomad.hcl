variable "image" {
  default = "csi-madrid:local" # or csi-madrid:local during development
}

variable "task_driver" {
  default = "podman" # could use podman
}

job "csi-madrid" {
  group "g" {
    task "t" {
      # inform Nomad that this task is a CSI plugin
      csi_plugin {
        id        = "madrid"
        type      = "monolith" # could also be run as separate "contoller" and "node" plugins
        mount_dir = "/csi"     # plugin listens on csi.sock in here
      }

      driver = var.task_driver
      config {
        image      = var.image
        privileged = true # node plugins in particular are usually privileged
      }

      # plugin code expects these env vars
      env {
        CSI_ENDPOINT = "/csi/csi.sock" # TODO: ${CSI_ENDPOINT}?
        NODE_ID      = "${node.unique.id}"

        # excluding these will cause the plugin to use an in-memory store instead,
        # which would result in volume state being lost on plugin restart.
        NOMAD_ADDR      = "unix:/secrets/api.sock"
        NOMAD_SINK_PATH = "csi-madrid" # matches the Nomad var path in policy.hcl
      }

      # we'll use Nomad's task API to store volume/snapshot state in variables
      identity {
        env = true # exposes $NOMAD_TOKEN and api.sock
      }
    }
  }
}
