# nfs_exporter
NFS exporter for Prometheus

## Installation

```
go get -u -v github.com/wangzewang/nfs_exporter

./${GOPATH}/bin/nfs_exporter --${flags} ...
```


## Usage of `nfs_exporter`

| Option                    | Default             | Description
| ------------------------- | ------------------- | -----------------
| -h, --help                | -                   | Displays usage.
| --web.listen-address      | `:9102`             | The address to listen on for HTTP requests.
| --web.metrics-path        | `/metrics`          | URL Endpoint for metrics
| --nfs.storage-path        | `/tmp/nfs`          | The nfs storage mount path
| --nfs.address             | `127.0.0.1`         | The nfs server IP address
| --version                 | -                   | Prints version information


## Reference
[https://github.com/aixeshunter/nfs_exporter](https://github.com/aixeshunter/nfs_exporter)
