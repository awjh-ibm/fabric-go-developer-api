# Fabric developer API
This project provides the contract interface, a high level API for application developers to implement business logic for Hyperledger Fabric.

## Tutorial
This tutorial will assume that you have the following prerequisites:
- Go

And in your Gopath:
- github.com/hyperledger/fabric
- github.com/hyperledger/fabric-samples
- github.com/awjh-ibm/fabric-go-developer-api

### Housekeeping
You will need to configure the docker containers used by fabric-samples/chaincode-docker-devmode to include this repository. To do this add the following line to the bottom of the `volumes` sections of `chaincode` and `cli` in `fabric-samples/chaincode-docker-devmode/docker-compose-simple.yaml`:

```
- $GOPATH/src/github.com/awjh-ibm/fabric-go-developer-api:/opt/gopath/src/github.com/awjh-ibm/fabric-go-developer-api
```

The folder chaincode within fabric-samples is already configured to be copied across to the docker container. Create a new folder within here called `go-developer-api-tutorial`. This is where you will write your tutorial chaincode.

### A simple contract
### Decalring a contract
The contractapi generates chaincode by taking one or more "contracts" that it bundles into a running chaincode. The first thing we will do here is declare a contract for use in our chaincode. This contract will be simple, handling the reading and writing of strings to and from the world state. All contracts for use in chaincode must implement the contractapi.ContractInterface. The easiest way to do this is embed the contractapi.Contract struct within your own contract which will provide default functionality for meeting this interface. 

Begin your contract by creating a folder `vendor` within `go-developer-api-tutorial` and adding a further folder `contracts` within the `vendor` folder. Create a file in `go-developer-api-tutorial/vendor/contracts` called `simple.go`. Within this file create a struct called Simple which embeds the contractapi.Contract struct:

```
package contracts

import (
	"github.com/hyperledger/fabric/core/chaincode/contractapi"
)

type Simple struct {
    contractapi.Contract
}
```

### Writing contract functions

All public functions of a struct are assumed to be callable via the final chaincode, these much match a specific format. If a public function does not match the format a panic will occur when the contract is used to create a chaincode. Functions of contracts for use in chaincode may take zero or more arguments and may return zero, one or two values.

The following are permissible types that may be taken in:
- *contractapi.TransactionContext (or a custom transaction context implementing contractapi.TransactionContextInterface)
- basic Go types:
    - string
    - bool
    - int (including int8, int16, int32 and int64)
    - uint (including uint8, uint16, uint32 and uint64)
    - float32
    - float64
- an array of length > 0 of any of the basic types
- a slice of the basic types
- multidimensional slice/array/combination of slice/array of basic types

If the function takes in *contractapi.TransactionContext (or a custom transaction context implementing contractapi.TransactionContextInterface) then that argument must be specified first within the function declaration and there may be only zero or one arguments of this type. There may be any number of the other types.

As values are passed as strings to fabric the contractapi will convert these to the correct go type. In cases of numeric types the conversion assumes base 10. If you wish to use another base take a string type and perform the conversion manually. Bool uses strconv.ParseBool. Arrays/Slices are assumed to be received in stringified JSON format. If conversion fails for any of the arguments passed an error is returned to the peer and the function is not called.

Functions can be defined to return zero, one or two values. These can be of types:
- string
- bool
- int (including int8, int16, int32 and int64)
- uint (including uint8, uint16, uint32 and uint64)
- float32
- float64
- an array of length > 0 of any of the above types
- a slice of the above types
- multidimensional slice/array/combination of slice/array of the above types
- error

At most one non-error type and one error can be returned. By go convention an error is expected as the last return type specified in the function declaration. The peer will receive the following responses (providing there are no errors elsewhere e.g. type conversion):
- **No return type specified** - peer receives a *success* response with no message for every function call
- **Single non error return type specified**  - returned value will be returned to the peer as a *success* with the returned value in the success message.
- **Single error return type specified** - if the returned error value is nil then a *success* response is returned to the peer with no message. If the error value is not nil then an *error* response is returned to the peer with the returned error message.
- **Two return types specified** - if the error return value is not nil then an *error* response is returned to the peer with the returned error message. If the error value is nil then a *success* response is sent to the peer with a stringified version of the returned value. For array/slice types a JSON format is used, for other non string types fmt.Sprintf is used to convert the value. 

The first function we will write for our contract is Create

-- TODO ADD FUNCTIONS FOR SIMPLE CONTRACT
-- EXTEND SIMPLE CONTRACT TO USE CONTRACTAPI PROVIDED FUNCTIONS
-- DO MORE COMPLEX ASSET E.G. TAKE IN NUMBERS