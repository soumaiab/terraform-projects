output "hello_status" {
  value = helm_release.hello.status
}

output "hello_go_service" {
  value = {
    name      = kubernetes_service.hello_go.metadata[0].name
    namespace = kubernetes_service.hello_go.metadata[0].namespace
  }
}
