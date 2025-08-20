resource "helm_release" "hello" {
  name       = "hello-world-softonic"
  repository = "https://charts.softonic.io"
  chart      = "hello-world-app"
  version    = "1.2.2"
}
