variable "image" {
  # use "csi-madrid:local" during development (from `make image` in repo root)
  default = "ghcr.io/gulducat/csi-madrid:main"
}

variable "task_driver" {
  default = "docker" # could use podman with appropriate agent config
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
        image = var.image
        args = [
          "-csi-endpoint=/csi/csi.sock", # TODO: ${CSI_ENDPOINT}?
          "-log-level=debug",
          "-node-id=${node.unique.id}",
          "-sink-nomad-path=csi-madrid", # matches the Nomad var path in policy.hcl
          # or can save volume/snapshot state to a file, like
          # "-sink-file-path=/tmp/somewhere/"
          # or exclude -sink-* to use an in-memory store.
        ]
        privileged = true # node plugins in particular are usually privileged
      }

      # we'll use Nomad's task API to store volume/snapshot state in variables
      identity {
        env = true # exposes $NOMAD_TOKEN and api.sock
      }
      env {
        NOMAD_ADDR = "unix:/secrets/api.sock" # TODO: flag?
      }
    }

    # speed up any test failures
    restart { attempts = 0 }
    reschedule {
      attempts  = 0
      unlimited = false
    }
  }
}
