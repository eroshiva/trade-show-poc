# This is inspired by:
# https://github.com/onosproject/onos-e2t/blob/master/build/onos-e2t/Dockerfile

# this is a minimal image, which needs git (and curl) to be installed
FROM golang:1.24-bookworm AS builder

# installing dependencies
RUN apt-get update && apt-get install -y --no-install-recommends make build-essential git && rm -rf /var/lib/apt/lists/*

# building a binary
ENV GO111MODULE=on
ENV GOPRIVATE=github.com/eroshiva
ARG MAKE_TARGET=build-simulator

COPY . /simulator

WORKDIR /simulator

RUN make ${MAKE_TARGET}

# building a small image
FROM gcr.io/distroless/base-debian12:nonroot

# Copy our static executable
COPY --from=builder /simulator/build/_output/nd-simulator /usr/local/bin/nd-simulator

# Set the entrypoint for the container
ENTRYPOINT ["/usr/local/bin/nd-simulator"]
