namespace "default" {
  variables {
    path "csi-madrid" {
      capabilities = ["write", "read", "list", "destroy"]
    }
    path "csi-madrid/*" {
      capabilities = ["write", "read", "list", "destroy"]
    }
  }
}
