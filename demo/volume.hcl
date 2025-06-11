id        = "my-vol"
name      = "my-vol"
type      = "csi"
plugin_id = "madrid"

capacity_max = "15mb"
capacity_min = "15mb"

secrets = {
  sssh = "be vewwy vewwy quiet"
}

capability {
  access_mode     = "multi-node-multi-writer"
  attachment_mode = "file-system"
}
capability {
  access_mode     = "multi-node-single-writer"
  attachment_mode = "file-system"
}
capability {
  access_mode     = "multi-node-reader-only"
  attachment_mode = "file-system"
}
