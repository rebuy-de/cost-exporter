# cost-exporter

![Build Status](https://github.com/rebuy-de/cost-exporter/workflows/Golang%20CI/badge.svg?branch=master)
![license](https://img.shields.io/github/license/rebuy-de/cost-exporter.svg)

Retrieves cost metrics and core counts from the AWS API and exposes this information via a Prometheus `/metrics` endpoint.

> **Development Status** *cost-exporter* is for internal use only. Feel free to use
> it, but expect big unanounced changes at any time. Furthermore we are very
> restrictive about code changes, therefore you should communicate any changes
> before working on an issue.


## Installation

Docker containers are are provided [here](https://quay.io/repository/rebuy/cost-exporter). To obtain the latest docker image run `docker pull quay.io/rebuy/cost-exporter:master`.

To compile *cost-exporter* from source you need a working
[Golang](https://golang.org/doc/install) development environment.

Then you just need to run `./buildutil` to compile a binary into the project
directory which you can then execute. With `./buildutil -x linux/arm64`
you can cross compile *cost-exporter* for other platforms.


## Usage

**cost-exporter**'s configuration is done using an configuration file that is pointed to when running the command, as well as a flag that defines the port it should listen on:
```
cost-exporter --config=/cost-exporter/config.yaml --port=8080
```

For more information, run:
```
cost-exporter --help
```


### Running in Kubernetes

Please see the `example/k8s/` directory for an example of how to run `cost-exporter` in Kubernetes.
