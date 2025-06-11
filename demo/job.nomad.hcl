variable "task_driver" {
  default = "docker"
}

job "web" {
  group "g" {
    # use the volume!
    volume "madrid" {
      type            = "csi"
      source          = "my-vol"
      attachment_mode = "file-system"
      access_mode     = "single-node-writer"
    }

    task "t" {
      driver = var.task_driver
      config {
        image   = "python:slim"
        command = "python"
        args    = ["-m", "http.server", "--directory=/madrid", "${NOMAD_PORT_http}"]
      }
      # mount the volume, too
      volume_mount {
        volume      = "madrid"
        destination = "/madrid"
      }
    }

    service {
      name     = "web"
      port     = "http"
      provider = "nomad"
    }
    network {
      port "http" {
        static = 8000
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
