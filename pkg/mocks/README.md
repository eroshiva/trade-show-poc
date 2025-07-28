# Network Device Simulator
This package implements simple Network Device Simulator suitable for testing this microservice.
It is solely gRPC-based, connectors package does not implement anything else than pure gRPC protocol.

Network Device Simulator return variables can be parametrised by setting environmental variables (especially, inside
 the Docker image).
> Checksum of the SW and FW version is always a SHA256 hash of a version.
