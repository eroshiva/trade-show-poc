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
    - Checksum (of the HW and SW, I guess?)

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
