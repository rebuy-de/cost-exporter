# cost-exporter

[![Build Status](https://travis-ci.org/rebuy-de/cost-exporter.svg?branch=master)](https://travis-ci.org/rebuy-de/cost-exporter)
[![license](https://img.shields.io/github/license/rebuy-de/cost-exporter.svg)]()

Retrieves cost metrics and core counts from the AWS API and exposes this information via a Prometheus `/metrics` endpoint.

> **Development Status** *cost-exporter* is for internal use only. Feel free to use
> it, but expect big unanounced changes at any time. Furthermore we are very
> restrictive about code changes, therefore you should communicate any changes
> before working on an issue.


## Installation

Docker containers are are provided [here](https://quay.io/repository/rebuy/cost-exporter). To obtain the latest docker image run `docker pull quay.io/rebuy/cost-exporter:master`.

To compile *cost-exporter* from source you need a working
[Golang](https://golang.org/doc/install) development environment. The sources
must be cloned to `$GOPATH/src/github.com/rebuy-de/cost-exporter`.

Also you need to install [godep](github.com/golang/dep/cmd/dep),
[golint](https://github.com/golang/lint/) and [GNU
Make](https://www.gnu.org/software/make/).

Then you just need to run `make build` to compile a binary into the project
directory or `make install` to install *cost-exporter* into `$GOPATH/bin`. With
`make xc` you can cross compile *cost-exporter* for other platforms.


## Usage

**node-drainer**'s configuration is done using an configuration file that is pointed to when running the command, as well as a flag that defines the port it should listen on:
```
cost-exporter --config=/cost-exporter/config.yaml --port=8080
```

For more information, run:
```
cost-exporter --help
```


### Running in Kubernetes

Please see the `example/k8s/` directory for an example of how to run `cost-exporter` in Kubernetes.
