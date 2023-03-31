# integrity-sum
[![GitHub contributors](https://img.shields.io/github/contributors/ScienceSoft-Inc/integrity-sum)](https://github.com/ScienceSoft-Inc/integrity-sum)
[![GitHub last commit](https://img.shields.io/github/last-commit/ScienceSoft-Inc/integrity-sum)](https://github.com/ScienceSoft-Inc/integrity-sum)
[![GitHub](https://img.shields.io/github/license/ScienceSoft-Inc/integrity-sum)](https://github.com/ScienceSoft-Inc/integrity-sum/blob/main/LICENSE)
[![GitHub issues](https://img.shields.io/github/issues/ScienceSoft-Inc/integrity-sum)](https://github.com/ScienceSoft-Inc/integrity-sum/issues)
[![GitHub forks](https://img.shields.io/github/forks/ScienceSoft-Inc/integrity-sum)](https://github.com/ScienceSoft-Inc/integrity-sum/network/members)

![Go](https://img.shields.io/badge/go-%2300ADD8.svg?style=for-the-badge&logo=go&logoColor=white)
![Kubernetes](https://img.shields.io/badge/kubernetes-%23326ce5.svg?style=for-the-badge&logo=kubernetes&logoColor=white)
![Postgres](https://img.shields.io/badge/postgres-%23316192.svg?style=for-the-badge&logo=postgresql&logoColor=white)
![GitHub](https://img.shields.io/badge/github-%23121011.svg?style=for-the-badge&logo=github&logoColor=white)

This program provides integrity monitoring that checks file or directory of container to determine whether or not they have been tampered with or corrupted.
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
### Statechart diagram
![File location: docs/diagrams/integrityStatechartDiagram.png](/docs/diagrams/integrityStatechartDiagram.png?raw=true "Statechart diagram")

## Getting Started
### Clone repository and install dependencies
```
$ cd path/to/install
$ git clone https://github.com/ScienceSoft-Inc/integrity-sum.git
```
Download the named modules into the module cache
```
go mod download
```

## Demo-App
You can test this application in your CLI — Command Line Interface on local files and folders.
You can use it with option(flags) like:
1) **`-d`** (path to dir):
```
go run cmd/demo-app/main.go -d ./..
```
2) **`-a`** (hash algorithm):
```
go run cmd/demo-app/main.go -a sha256
go run cmd/demo-app/main.go -a SHA256
go run cmd/demo-app/main.go -a SHA256 -d ./..
```
3) **`-h`** (options docs):
```
go run cmd/demo-app/main.go -h
```

## :hammer: Installing components
### Running locally
The code only works running inside a pod in Kubernetes.
You need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster.
If you do not already have a cluster, you can create one by using `minikube`.
Example https://minikube.sigs.k8s.io/docs/start/

### Install Helm
Before using helm charts you need to install helm on your local machine.
You can find the necessary installation information at this link https://helm.sh/docs/intro/install/

### Configuration
To work properly, you first need to set the configuration files:
+ environmental variables in the `.env` file
+ values in the file `helm-charts/database-to-integrity-sum/values.yaml`
+ values in the file `helm-charts/app-to-monitor/values.yaml`

## Quick start
### Using Makefile
You can use make function.
Runs all necessary cleaning targets and dependencies for the project:
```
make all
```
Remove an installed Helm deployments and stop minikube:
```
make stop
```
Building and running the project on a local machine:
```
make run
```
If you want to generate binaries for different platforms:
```
make compile
```
### Manual start
First of all, get a build image with `make buildtools`. It will be used to compile source code (Go & C/C++).

optional:
- SYSLOG_ENABLED=true, if syslog functionality is required;
- BEE2_ENABLED=true, if bee2 hash algorithm is required.

Build app: `make build`

Minikube start:
```
minikube start
```
Build docker image:
```
make docker
```
If you use `kind` instead of `minikube`, you may do `make kind-load-images` to preload image.

This command installs a chart archive.
```
helm install `release name` `path to a packaged chart`
```
There are some predefined targets in the Makefile for deployment:

- Install helm chart with app
    ```
    make helm-app
    ```

## Pay attention!
If you want to use a hasher-sidecar, then you need to specify the following data in your deployment:
Pod annotations:
+ `integrity-monitor.scnsoft.com/inject: "true"` - The sidecar injection annotation. If true, sidecar will be injected.
+ `<monitoring process name>.integrity-monitor.scnsoft.com/monitoring-paths: etc/nginx,usr/bin` - This annotation introduces a process to be monitored and specifies its paths.
Annotation prefix should be the process name for instance `nginx`, `nginx.integrity-monitor.scnsoft.com/monitoring-paths`
Service account:
+ `template:spec:serviceAccountName:` api-version-`hasher`
Share process namespace should be enabled.
+ `template:shareProcessNamespace: true`

## Troubleshooting
Sometimes you may find that pod is injected with sidecar container as expected, check the following items:

1) The pod is in running state with `integrity` sidecar container injected and no error logs.
2) Check if the application pod has the correct annotations as described above.
___________________________
## :notebook_with_decorative_cover: Godoc extracts and generates documentation for Go programs
#### Presents the documentation as a web page.
```go
godoc -http=:6060/integritySum
go doc packge.function_name
```
for example
```go
go doc pkg/api.Result
```

## :mag: Running tests
First of all you need to install mockgen:
```
go install github.com/golang/mock/mockgen@${VERSION_MOCKGEN}
```

Generate a mock:
```
go generate ./internal/core/ports/repository.go
go generate ./internal/core/ports/service.go
```
or
```
make generate
```
You need to go to the folder where the file is located *_test.go and run the following command:
```go
go test -v
```

for example
```go
cd ../pkg/api
go test -v
```
or
```
go test -v ./...
```
or
```
make test
```
## :mag: Running linter "golangci-lint"
```
golangci-lint run
```

## Including Bee2 library into the application

Use the Makefile target `bee2-lib` to build standalone static and shared binary for the bee2 library.
Use the env variable `BEE2_ENABLED=true` with `make build` to include bee2 library into the application. Then deployment may be updated with `--algorithm=BEE2` arg to select bee2 hashing algorithm.

Find more details about bee2 tools in the [Readme](internal/ffi/bee2/Readme.md).

## Enable MinIO

The code was tested with default `bitnami/minio` helm chart.

### Install standalone server

The following code will create the `minio` namespace and install a default MinIO server into it.

```bash
kubectl create ns minio
helm install minio --namespace=minio bitnami/minio
```

Refer to the original documentation [Bitnami Object Storage based on MinIO®](https://hub.docker.com/r/bitnami/minio/) for more details.

### Include into the project

To enable the MinIO in the project set the `minio.enabled` to `true` in the `helm-charts/app-to-monitor/values.yaml`.

## Syslog support

In order to enable syslog functionality following flags should be set:

- --syslog-enabled=true
- --syslog-host=syslog.host
- --syslog-port=syslog.port default 514
- --syslog-proto=syslog protocol default TCP

It could be done either through deployment or by integrity injector.

To enable syslog support in demo-app:

- Set `SYSLOG_ENABLED` environment variable to `true`
- Install demo app by running: 
```
make helm-app
``` 

### Install syslog server

Syslog helm chart is included in demo app, to install it following steps should be done:

- Set `SYSLOG_ENABLED` environment variable to `true`
- Run: ```make helm-syslog```

### Syslog messages format

Syslog message format conforms to rfc3164 https://www.ietf.org/rfc/rfc3164.txt
e.g.

```
 <PRI> TIMESTAMP HOSTNAME TAG time=<event timestamp> event-type=<00001> service=<service name> namespace=<namespace name> cluster=<cluster name> message=<event message> file=<changed file name> reason=<event reason>\n`
```

- PRI - message priority, always 28 (LOG_WARNING | LOG_DAEMON)
- TIMESTAMP - time stamp format  "Jan _2 15:04:05"
- HOSTNAME - host/pod name
- TAG - process name with pid e.g. integrity-monitor[2]:

USER message consists from key=value pairs

- time=\<event timestamp\> , time stamp format  "Jan _2 15:04:05"
- event-type=\<00001\>
  - `00001` - "file content mismatch"
  - `00002` - "new file found"
  - `00003` - "file deleted"
- service=\<service name\>, monitoring service name e.g. `service=nginx`
- namespace=\<namespace name\>, pod namespace
- cluster=\<cluster name\>, service cluster name
- message=\<event message\>, e.g. `message=Restart deployment`
- file=\<changed file name\>, full file name with changes detected
- reason=\<event reason\>, restart reason e.g.:
  - `file content mismatch`
  - `new file found`
  - `file deleted`

Message examples from syslog:

```log
Mar 20 13:56:17 nginx-webhook-c74fc7b97-92n7k integrity-monitor[69]: time=Mar 20 13:56:17 event-type=0001 service=nginx namespace=default cluster=dev message=Restart deployment nginx-webhook file=/proc/46/root/etc/nginx/nginx.conf reason=file content mismatch
Mar 20 14:08:50 nginx-webhook-755ff64f94-5bclh integrity-monitor[47]: time=Mar 20 14:08:50 event-type=0001 service=nginx namespace=default cluster=dev message=Restart deployment nginx-webhook file=/proc/46/root/etc/nginx/qqqqq reason=new file found
Mar 20 14:10:23 nginx-webhook-86f7cf989c-5pszt integrity-monitor[47]: time=Mar 20 14:10:23 event-type=0001 service=nginx namespace=default cluster=dev message=Restart deployment nginx-webhook file=/proc/46/root/etc/nginx/nginx.conf reason=file deleted
```

Note: `<PRI>` - is not shown up in syslog server logs.

## License

This project uses the MIT software license. See [full license file](https://github.com/ScienceSoft-Inc/integrity-sum/blob/main/LICENSE)
