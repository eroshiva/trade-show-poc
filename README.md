# Trade Show PoC
This is an assessment task for the Ubiquiti interview process.
[Here](./task_definition.md) is a task definition based on the PDF description with my comments.


## Development prerequisites
To install development prerequisites, run `make deps`. It will install all necessary plugins for code generation.


## Running the demo
To run the demo, simply run `make poc` command. It will create local Kubernetes cluster (with `kind`), build Golang code
and Docker images, upload those images to Kubernetes cluster, and deploy the monitoring solution with `helm` charts.

There is also a set of simulators to showcase the merits of the system and ability to interact with network devices with 
different protocols.

Additionally, a simple CLI tool was developed to enable interaction between gRPC and REST endpoints of the monitoring
solution.

> In the perfect world, where there is more time, I'd love to extend the solution with Grafana visualisation or other 
> similar solutions.


## Solution
This section describes a provided solution.


## Testing
This section describes testing procedure.


## What can be done better
ABAC access control to the resources should be implemented to better restrict access to the fields of the resources 
in the data schema, namely:
- User can update only network device model, vendor, and endpoints.
- Controller itself can retrieve and update only network device HW, SW, FW, and device status.
- More sanity checks on the input data must be added at the API (gRPC server) side and at the DB client side.



## Disclaimer
This section gives an honest opinion on the development process, in particular on the use of AI tools.


### Use of AI tools
AI tools were not used for any code generation neither code completion nor for coding instead of me. Somehow,
Gemini 2.5 Pro was used to make initial research in best practices for handling:
- REST and gRPC API simultaneously.
    - Previous idea was to have two API Gateways - gRPC and REST one, but too much work.
    - Research indicated the existence of the grpc-gateway plugin, which is able to autogenerate rever HTTP
      proxy out of Protobuf definition of the schema.
- SQL coexistance with Go code.
    - Research in tooling — rather misleading, unhelpful, and time-consuming.
    - I had to stick with my original idea to use `protoc-gen-ent` and `ent` framework for PostgreSQL interaction with
      microservice, which provided a central place for managing everything - API and SQL-driven schema within a single Protobuf.

I also found some recommendations about different tool usage confusing and misleading rather than helpful.
It's always better to follow tool's documentation rather than asking AI for a tutorial.


### Silly, nasty bugs
This project has been evolving rapidly fast, under the short time constraints. There might be some little tricky corner 
cases that I didn't take into account. Overall, the general workflow should be safe from violations. In case you found a bug,
feel free to submit an issue with a detailed description. 
