# Network Device Simulator
This package implements a simple Network Device Simulator suitable for testing this microservice.
It is solely gRPC-based and `connectors` package does not implement anything else than pure gRPC protocol connectivity.

You can also find a Dockerfile for wrapping this simulator into a container, which can ran in cloud environment for the PoC 
testing.

Network Device Simulator return variables can be parametrised by setting environmental variables (especially, inside 
of the Docker image). For the reference, please see constants specified on top of the [simulator.go](./simulator.go) file.
> Checksum of the SW and FW version is always a SHA256 hash of a version.
