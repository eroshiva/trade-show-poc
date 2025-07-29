## External checksum generator
Within this package an embedding of `external binary checksum generator` has been introduced.
Main idea is to invoke it from Golang code through command line package and then parse the output (assuming it 
produces only checksum).

For the sake of testability, an interface and a mock implementation of this interface were created. Mocked checksum
generator is then embedded from the `main()` function to the `manager`'s main control loop.
Mock checksum generator is kept simple - it creates a SHA256 checksum on the SW/FW version string that is provided at 
its input (the same way checksums are generated in [Network Device Simulator](../mocks/README.md)).