resource "helm_release" "hello" {
  name       = "hello-world-softonic"
  repository = "https://charts.softonic.io"
  chart      = "hello-world-app"
  version    = "1.2.2"
}

# Optional: isolate in its own namespace
resource "kubernetes_namespace" "hello" {
  metadata { name = "hello" }
}

resource "kubernetes_deployment" "hello_go" {
  metadata {
    name      = "hello-go"
    namespace = kubernetes_namespace.hello.metadata[0].name
    labels = { app = "hello-go" }
  }

  spec {
    replicas = 1
    selector { match_labels = { app = "hello-go" } }

    template {
      metadata { labels = { app = "hello-go" } }

      spec {
        container {
          name  = "server"
          image = "hellogo:0.3"         # built into Minikube
          image_pull_policy = "IfNotPresent"

          port { container_port = 8080 }

          env {
            name  = "APP_MESSAGE"
            value = var.app_message      # <- comes from Terraform
          }
        }
      }
    }
  }
}

resource "kubernetes_service" "hello_go" {
  metadata {
    name      = "hello-go"
    namespace = kubernetes_namespace.hello.metadata[0].name
    labels    = { app = "hello-go" }
  }

  spec {
    selector = { app = "hello-go" }
    port {
      port        = 80
      target_port = 8080
    }
    type = "ClusterIP"
  }
}
