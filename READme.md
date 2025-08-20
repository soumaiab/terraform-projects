# Running the app:

## Setup
From Ubuntu:
> cd /mnt/c/Users/<USER>/<CODE_PATH>

> minikube start

> kubectl config use-context minikube

> tofu init -upgrade

> tofu plan

> tofu apply -auto-approve


## Port forwarding:

Check the name of service and the port:
> kubectl get pods -A
> kubectl get svc -A

> kubectl port-forward svc/<SERVICE_NAME (in my case: kubectl port-forward svc/hello-world-softonic-hello-world-app 8080:80)> <PORT>:80