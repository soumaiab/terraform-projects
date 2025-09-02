# Running the app:

## Setup
From Ubuntu:
> cd /mnt/c/Users/<USER_NAME>/<CODE_PATH>

> minikube start

> kubectl config use-context minikube

> tofu init -upgrade

> tofu plan

> tofu apply -auto-approve


## Port forwarding the Helm Chart:

Check the name of service and the port:
> kubectl get pods -A

> kubectl get svc -A

> kubectl port-forward svc/<SERVICE_NAME> <PORT_NUM>:80

In my case: `kubectl port-forward svc/hello-world-softonic-hello-world-app 8080:80`

# Testing the Go server
> go build -o server.exe main.go

> .\server.exe

It should connect to http://localhost:8080/.
Since it wasn't set by tofu, we should see (APP_MESSAGE not set).

# Building the image and running the app

On Ubuntu from the source folder:
> minikube image build -t hellogo:0.<TAG NUM> .

If anything changes in the go code, rebuild and increase the tag number. It is currently hellogo:0.3


Apply changes made to providers (only run if not done before/not changes):
> tofu init -upgrade

Apply the message
> tofu apply -auto-approve -var="app_message=Hello from OpenTofu!"

Ensure it got deployed:
> kubectl -n hello get deploy hello-go

> kubectl -n hello get pods -l app=hello-go

> kubectl -n hello logs deploy/hello-go

Should see: "Listening on :8080".

To run it:
> kubectl -n hello port-forward svc/hello-go 8080:80

To update the message:
> tofu apply -auto-approve -var="app_message=Hello Soumaia 🚀"


#### Stopping minikube:
> minikube stop

#### Restarting it

Start Minikube back up
> minikube start

Check that your namespace & deployment are still there
> kubectl get ns

> kubectl -n hello get deploy hello-go

> kubectl -n hello get pods -l app=hello-go

If you see a pod with STATUS=Running, the Go app is up again.

Port-forward the Service
> kubectl -n hello port-forward svc/hello-go 8080:80

If the pod isn’t running
> minikube image build -t hellogo:0.3 .

Restart deployment so it picks up the image
> kubectl -n hello rollout restart deploy/hello-go