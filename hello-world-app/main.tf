# resource "helm_release" "hello" {
#  name       = "hello-world-softonic"
#  repository = "https://charts.softonic.io"
#  chart      = "hello-world-app"
#  version    = "1.2.2"
# }

# Optional: isolate in its own namespace
resource "kubernetes_namespace" "hello" {
  metadata { name = "hello" }
}

# Secret for env
resource "kubernetes_secret" "hello_env" {
  metadata {
    name      = "hello-go-secret"
    namespace = kubernetes_namespace.hello.metadata[0].name
  }
  data = {
    APP_MESSAGE = var.app_message   
  }
}

# Replaces deployment + service and gets the values for those from values.yml
resource "helm_release" "hello_go" {
  repository       = "https://stakater.github.io/stakater-charts"
  chart            = "application"
  version          = "6.0.0"

  name             = "hello-go"
  namespace        = kubernetes_namespace.hello.metadata[0].name
  create_namespace = false

  values = [
    templatefile("${path.module}/values.yml", {
      # Injecting those values in the yml file
      env_secret_name  = kubernetes_secret.hello_env.metadata[0].name
      image_repository = "hellogo"   # your local Minikube image name
      image_tag        = "0.3"       # matches your Docker tag
    })
  ]
}

