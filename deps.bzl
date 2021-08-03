load("@bazel_gazelle//:deps.bzl", "go_repository")

def go_dependencies():
    go_repository(
        name = "io_etcd_go_bbolt",
        importpath = "go.etcd.io/bbolt",
        sum = "h1:XAzx9gjCb0Rxj7EoqcClPD1d5ZBxZJk0jbuoPHenBt0=",
        version = "v1.3.5",
    )
    go_repository(
        name = "org_golang_dl",
        importpath = "golang.org/dl",
        sum = "h1:FwHRCy3afe8UnKJCTzoA1feDrGENbl70KqGFVMu2y3Y=",
        version = "v0.0.0-20201217181409-aeefed14b4e2",
    )
    go_repository(
        name = "org_golang_x_sys",
        importpath = "golang.org/x/sys",
        sum = "h1:LfCXLvNmTYH9kEmVgqbnsWfruoXZIrh4YBgqVHtDvw0=",
        version = "v0.0.0-20200202164722-d101bd2416d5",
    )
