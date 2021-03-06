# nfs_exporter
NFS exporter for Prometheus

## Installation

```
git clone https://github.com/wangzewang/nfs_exporter.git $GOPATH/src/github.com/wangzewang/nfs_exporter

cd $GOPATH/src/github.com/wangzewang/nfs_exporter/

go mod vendor

CGO_ENABLED=0 go build .

docker build  --network=host .

# Push the docker image to your registry and change config in yaml folder
# then deploy nfs exporter to kubernetes

kubectl apply -f ./yaml
```

## Usage of `nfs_exporter`

| Option                    | Default             | Description
| ------------------------- | ------------------- | -----------------
| -h, --help                | -                   | Displays usage.
| --web.listen-address      | `:9102`             | The address to listen on for HTTP requests.
| --web.metrics-path        | `/metrics`          | URL Endpoint for metrics
| --nfs.address             | `127.0.0.1:/xxxxx, 127.0.0.2:/xxx`  | The nfs server IP Path address
| --version                 | -                   | Prints version information



## Reference
[https://github.com/aixeshunter/nfs_exporter](https://github.com/aixeshunter/nfs_exporter)
