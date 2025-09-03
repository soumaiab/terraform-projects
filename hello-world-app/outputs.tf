# output "hello_status" {
#   value = helm_release.hello.status
# }

# Look up the Service created by the Helm chart
data "kubernetes_service" "hello_go" {
  metadata {
    name      = "hello-go"                                # <- must match your values.yml applicationName
    namespace = kubernetes_namespace.hello.metadata[0].name
  }
}

output "hello_go_service" {
  value = {
    name      = data.kubernetes_service.hello_go.metadata[0].name
    namespace = data.kubernetes_service.hello_go.metadata[0].namespace
    status    = helm_release.hello_go.status
    version   = helm_release.hello_go.version
    cluster_ip = try(data.kubernetes_service.hello_go.spec[0].cluster_ip, null)
    ports      = try(data.kubernetes_service.hello_go.spec[0].port, [])
  }
}