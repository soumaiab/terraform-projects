Ran:

## Set up a plain Go project

mkdir hello-operator
cd hello-operator
go mod init example.com/hello-operator

go get k8s.io/client-go@latest
go get k8s.io/apimachinery@latest

## Step 2 — Write the CRD (helloapps.yaml)

Create a file called helloapps-crd.yaml

kubectl apply -f helloapps-crd.yaml

## Step 3 — Create a sample HelloApp

Make a file helloapp-sample.yaml

kubectl apply -f helloapp-sample.yaml

## Step 4 — Write a controller skeleton
