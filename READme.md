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

# Testing the Go server
> go build -o server.exe main.go
> .\server.exe

It should connect to http://localhost:8888/.
Since it wasn't set by tofu, we should see (APP_MESSAGE not set).
