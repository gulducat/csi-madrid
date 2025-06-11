job "web" {
  group "g" {
    volume "madrid" {
      type            = "csi"
      source          = "my-vol"
      attachment_mode = "file-system"
      access_mode     = "single-node-writer"
    }
    task "t" {
      driver = "podman"
      config {
        image   = "python:slim"
        command = "python"
        args    = ["-m", "http.server", "--directory=/madrid", "${NOMAD_PORT_http}"]
      }
      volume_mount {
        volume      = "madrid"
        destination = "/madrid"
      }
    }
    service {
      name = "web"
      port = "http"
      provider = "nomad"
    }
    network {
      mode = "bridge"
      port "http" {
        static = 8000
      }
    }
  }
}
