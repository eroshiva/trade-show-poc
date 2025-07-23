## Task definition
It is required to deliver a small PoC for network device monitoring, which checks
if a list of the devices is alive.

Network monitoring service needs to support:
- gRPC calls,
- REST calls.

> For customer facing API, I assume.

- Retrieve following diagnostic data depending on the protocol that device supports:
    - Hardware version
    - Software version
    - Firmware version
    - Device status (?)
    - Checksum ((separately) for the HW, SW, and FW, I guess?)

> I assume, that interface is required there to interact with different protocols.

Another statement is `use its health endpoint to get the capabilities`. I interpret it as follows:
device has few endpoints, which are known in prior, solution should sniff all endpoints and communicate
with the first that works.

An API for retrieving the latest status of all devices is needed. A list of devices might change dynamically, so
it should be easy to update.
> Obvious proposal is to insert a list of the devices in the API call each time.
>
> Another possible solution is to have a separate APi for uploading a list of the devices to monitor.
> This list will be saved to the PostgreSQL. Monitoring routine will retrieve a list of the devices each time and
> conduct the monitoring (e.g., once per defined period in time). Probably a better solution.

If device is down, we need to handle it either with:

- retries,
- logging,
- another suggestions.

> Combination of retries and logging is preferrred.

Some devices might be behind the unstable network, which makes it reasonable to do retries with exponential backoff.
False alarms are prohibited.

Device-related data, e.g., device identities (probably meant device model with its endpoints), status, and other relevant
details should be stored in the PostgreSQL.

Another requirement is to `integrate the server library for the devices` (probably meant this microservice) `with an 
external checksum generator binary executable` (What is it? Another microservice to conduct checksums against?).

Full PoC must be paired with PoC life-cycle to get valid test results (whatever it means). Testing the whole thing
must be easy.
> To orchestrate testing, a Makefile has been put in place.

## Goals
Given the letter from `Boss`, following are defined goals of this task

- Build a microservice that implements network device monitoring logic,
  - It should be able to dynamically update a list of network devices that are monitored.
  - User-friendly and/or developer-friendly (both, gRPC and REST) API is required (e.g., add a `/summary` call or other useful wrappers for statistics).
  - Define and implement a generic interface that allows to communicate with network devices over different network protocols.
  - Build core logic for network device monitoring that allows to avoid false alarms and provides real-time (or near real-time) monitoring.
    - Consider the case when devices are located in a place with low bandwidth and bad network connectivity.
  - Network device state should be stored in PostgreSQL.
  - Code structure should be developer-friendly (i.e., no mess in code, clear structure and documentation).
- Implement/Mock `checksum generator binary executable` with which microservice interacts.
  - Also, implement mocked devices to showcase interaction over interface with various device protocols.
- Ease of deployment.
  - Containerise everything.
  - Orchestrate deployment with Makefile or similar, as close to production environment as possible).
