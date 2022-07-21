# integrity-sum
![GitHub contributors](https://img.shields.io/github/contributors/ScienceSoft-Inc/k8s-container-integrity-monitor)
![GitHub last commit](https://img.shields.io/github/last-commit/ScienceSoft-Inc/k8s-container-integrity-monitor)
![GitHub](https://img.shields.io/github/license/ScienceSoft-Inc/k8s-container-integrity-monitor)
![GitHub issues](https://img.shields.io/github/issues/ScienceSoft-Inc/k8s-container-integrity-monitor)
![GitHub forks](https://img.shields.io/github/forks/ScienceSoft-Inc/k8s-container-integrity-monitor)

![Go](https://img.shields.io/badge/go-%2300ADD8.svg?style=for-the-badge&logo=go&logoColor=white)
![Kubernetes](https://img.shields.io/badge/kubernetes-%23326ce5.svg?style=for-the-badge&logo=kubernetes&logoColor=white)
![Postgres](https://img.shields.io/badge/postgres-%23316192.svg?style=for-the-badge&logo=postgresql&logoColor=white)
![GitHub](https://img.shields.io/badge/github-%23121011.svg?style=for-the-badge&logo=github&logoColor=white)

This program provides integrity monitoring that checks file or   directory of container to determine whether or not they have been tampered with or corrupted.  
integrity-sum, which is a type of change auditing, verifies and validates these files by comparing them to the stored data in the database.

If program detects that files have been altered, updated, added or compromised, it rolls back deployments to a previous version.

integrity-sum injects a `hasher-sidecar` to your pods as a sidecar container. 
`hasher-sidecar` the implementation of a hasher in golang, which calculates the checksum of files using different algorithms in kubernetes:
* MD5
* SHA256
* SHA1
* SHA224
* SHA384
* SHA512

## Architecture

## Getting Started

### Clone repository and install dependencies
```
$ git clone https://github.com/ScienceSoft-Inc/integrity-sum.git
$ cd path/to/install
```
### Running locally
The code only works running inside a pod in Kubernetes
You need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster.
If you do not already have a cluster, you can create one by using `minikube`.  
Example https://minikube.sigs.k8s.io/docs/start/

### Configuration
To work properly, you first need to set the configuration files:
+ environmental variables in the `.env` file
+ config in file `manifests/hasher/configMap.yaml`
+ secret for database `manifests/database/postgres-secret.yaml`


## :hammer: Installing components

### Installation DATABASE
Apply all annotations in directory "manifests/db/..":
```
kubectl apply -f manifests/database/postgres-db-pv.yaml
kubectl apply -f manifests/database/postgres-db-pvc.yaml
kubectl apply -f manifests/database/postgres-secret.yaml
kubectl apply -f manifests/database/postgres-db-deployment.yaml
kubectl apply -f manifests/database/postgres-db-service.yaml
```

## Quick start
Build docker images hasher:
```
eval $(minikube docker-env)
docker build -t hasher .
```

Apply hasher annotation:
```
kubectl apply -f manifests/hasher/service-account-hasher.yaml
kubectl apply -f manifests/hasher/configMap.yaml
```

See examples in manifests/hasher directory for how to add the `hasher-sidecar` sidecar-container to any pod, and the service account needed.
For example there is manifests/hasher/test-deployment.yaml DEPLOYMENT files:
```
kubectl apply -f manifests/hasher/test-deployment.yaml
```

##Pay attention!
If you want to use a hasher-sidecar, then you need to specify the following data in your deployment:
+ `main-process-name: "your main process name"`
+ `template:spec:serviceAccountName: hasher` 
+ `template:shareProcessNamespace: true` 

## Troubleshooting
Sometimes you may find that pod is injected with sidecar container as expected, check the following items:

1) The pod is in running state with `hasher-sidecar` sidecar container injected and no error logs.
2) Check if the application pod has he correct labels `hasher-webhook-injector-sidecar: "true"` and installed `hasher-webhook-process-name`.
___________________________
### :notebook_with_decorative_cover: Godoc extracts and generates documentation for Go programs
#### Presents the documentation as a web page.
```go
godoc -http=:6060/sha256sum
go doc packge.function_name
```
for example
```go
go doc pkg/api.Result
```

### :mag: Running tests

You need to go to the folder where the file is located *_test.go and run the following command:
```go
go test -v
```

for example
```go
cd ../pkg/api
go test -v
```

### :mag: Running linter "golangci-lint"
```
golangci-lint run
```

##License
This project uses the MIT software license. See [full license file](https://github.com/ScienceSoft-Inc/integrity-sum/blob/main/LICENSE)
