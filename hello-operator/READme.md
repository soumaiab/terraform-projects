# Operator User Guide

## Setup
The image should already exist.

Applying the yamls:
> kubectl apply -f env-secret.yaml

> kubectl apply -f helloapps-crd.yaml

> kubectl apply -f my-hello-app.yaml

Running the operator:

> go run main.go 

## Modifications:
If modify the yml:

> kubectl apply -f <FNAME.yaml>


If we want to change the secret:

Change in the file itself then:
> kubectl apply -f env-secret.yaml

> kubectl rollout restart deploy my-hello

## Running the app (locally)
Option 1 — Port-forward:
> kubectl port-forward svc/my-hello 8080:80

Open http://localhost:8080/

OR

Option 2 — Minikube service:
> minikube service hello-go -n hello --url