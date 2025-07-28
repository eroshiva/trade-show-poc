## Short note from author
Normally, each connector should implement protocol-specific invocation. For the sake of simplicity and showcasing the 
testability, we are implementing everywhere, under the hood, pure gRPC connection instantiation.
This way, connectors from this package would be able to connect to the [Device Simulator](../mocks/README.md) and
retrieve necessary information.

### Things to consider in the future
Currently, new connection is instantiated at each call, which is not efficient and scalable in the long-term.
This should be reconsidered. 
- A possible way to approach it is to have an API gateway, which maintains pool of different 
connections (per protocol), and each request from monitoring microservice goes to this API gateway.
  - Favourable approach.
  - This way complexity of dealing with various protocols is moving to the other component.
- Another approach is to move connection instantiation one level above (to `manager`'s main control loop), but that
would overcomplicate the code and make it hard to maintain in the long-term future.
  - Least favourable approach.
  - Complexity of dealing with various protocols stays and makes code a bit of spaghetti with additional sauce.
- Another possible approach - a compromise between the other two aproaches - is to split device monitoring service per 
supported protocols. This way, common libraries could be reshared, making bootstrap of a microservice fast, and 
protocol specific implementations can go inside the microservice.
  - Meh approach.
  - Potentially too much work and maintenance.
