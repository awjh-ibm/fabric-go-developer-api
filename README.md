> WARNING: This is a prototype. This project is not intended for production use (yet). The documentation and tutorials may not match with the code. There is no branching all new code goes into master.

# Fabric developer API
This project provides the contract interface, a high level API for application developers to implement business logic for Hyperledger Fabric.

[Link to GoDoc](https://godoc.org/github.com/awjh-ibm/fabric-go-developer-api/contractapi)

## Tutorial
The tutorial for this API is split into three main parts:
- [A simple contract](./tutorials/a_simple_contract.md)
    - [Prerequisites](./tutorials/a_simple_contract.md#prerequisites)
    - [Housekeeping](./tutorials/a_simple_contract.md#housekeeping)
    - [Declaring a contract](./tutorials/a_simple_contract.md#declaring-a-contract)
    - [Writing contract functions](./tutorials/a_simple_contract.md#writing-contract-functions)
    - [Using contracts in chaincode](./tutorials/a_simple_contract.md#using-contracts-in-chaincode)
    - [Running your chaincode as a developer](./tutorials/a_simple_contract.md#running-your-chaincode-as-a-developer)
    - [Interacting with your running chaincode](./tutorials/a_simple_contract.md#interacting-with-your-running-chaincode)
- [Extending a simple contract](./tutorials/extending_a_simple_contract.md)
    - [Prerequisites](./tutorials/extending_a_simple_contract.md#prerequisites)
    - [Calling functions every time a request is made and using custom transaction contexts](./tutorials/extending_a_simple_contract.md#calling-functions-every-time-a-request-is-made-and-using-custom-transaction-contexts)
    - [Handling unknown function requests](./tutorials/extending_a_simple_contract.md#handling-unknown-function-requests)
    - [Interacting with your running chaincode](./tutorials/extending_a_simple_contract.md#interacting-with-your-running-chaincode)
- [Incorporating multiple contracts and handling objects](./tutorials/incorporating_multiple_contracts_and_handling_objects.md)
    - [Prerequisites](./tutorials/incorporating_multiple_contracts_and_handling_objects.md#prerequisites)
    - [Building a contract to handle an asset](./tutorials/incorporating_multiple_contracts_and_handling_objects.md#building-a-contract-to-handle-an-asset)
    - [Adding a second contract to the chaincode](./tutorials/incorporating_multiple_contracts_and_handling_objects.md#adding-a-second-contract-to-the-chaincode)
    - [Interacting with your running chaincode](./tutorials/incorporating_multiple_contracts_and_handling_objects.md#interacting-with-your-running-chaincode)

These follow on from each other so it is recommended you follow them in order.