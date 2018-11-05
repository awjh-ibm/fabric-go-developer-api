# A simple contract

- [Prerequisites](#prerequisites)
- [Housekeeping](#housekeeping)
- [Declaring a contract](#declaring-a-contract)
- [Writing contract functions](#writing-contract-functions)
- [Using contracts in chaincode](#using-contracts-in-chaincode)
- [Running your chaincode as a developer](#running-your-chaincode-as-a-developer)
- [Interacting with your running chaincode](#interacting-with-your-running-chaincode)

## Prerequisites
This tutorial will assume that you have the following prerequisites:
- Go
- Docker
- Docker compose

And in your Gopath:
- github.com/hyperledger/fabric
- github.com/hyperledger/fabric-samples
- github.com/awjh-ibm/fabric-go-developer-api

## Housekeeping
You will need to configure the docker containers used by fabric-samples/chaincode-docker-devmode to include this repository. To do this add the following line to the bottom of the `volumes` sections of `chaincode` and `cli` in `fabric-samples/chaincode-docker-devmode/docker-compose-simple.yaml`:

```
- $GOPATH/src/github.com/awjh-ibm/fabric-go-developer-api:/opt/gopath/src/github.com/awjh-ibm/fabric-go-developer-api
```

The folder chaincode within fabric-samples is already configured to be copied across to the docker container. Create a new folder within here called `go-developer-api-tutorial`. This is where you will write your tutorial chaincode.

## Declaring a contract
The contractapi generates chaincode by taking one or more "contracts" that it bundles into a running chaincode. The first thing we will do here is declare a contract for use in our chaincode. This contract will be simple, handling the reading and writing of strings to and from the world state. All contracts for use in chaincode must implement the contractapi.ContractInterface. The easiest way to do this is embed the contractapi.Contract struct within your own contract which will provide default functionality for meeting this interface. 

Begin your contract by creating a folder `vendor` within `go-developer-api-tutorial` and adding a further folder `contracts` within the `vendor` folder. Create a file in `go-developer-api-tutorial/vendor/contracts` called `simple.go`. Within this file create a struct called Simple which embeds the contractapi.Contract struct:

```
package contracts

import (
    "errors"
    "fmt"

    "github.com/awjh-ibm/fabric-go-developer-api/contractapi"
)

// Simple contract for handling writing and reading from the world state
type Simple struct {
    contractapi.Contract
}
```

## Writing contract functions

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

The first function to write for your contract is `Create`. This will add a new key value pair to the world state using a key and value provided by the user. As it interacts with the world state it will also require the transaction context to be passed. As the default transaction context provided by contractapi provides all the necessary functions for interacting with the world state the function will take this. As the function is intended to write rather than return data it will only return the error type.

```
// Create adds a new key with value to the world state
func (s *Simple) Create(ctx *contractapi.TransactionContext, key string, value string) error {
    existing, err := ctx.GetStub().GetState(key)

    if err != nil {
        return errors.New("Unable to interact with world state")
    }

    if existing != nil {
        return fmt.Errorf("Cannot create world state pair with key %s. Already exists", key)
    }

    err = ctx.GetStub().PutState(key, []byte(value))

    if err != nil {
        return errors.New("Unable to interact with world state")
    }

    return nil
}
```

The function uses the stub of the transaction context ([shim.ChaincodeStubInterface](https://godoc.org/github.com/hyperledger/fabric/core/chaincode/shim#ChaincodeStubInterface)) to first read from the world state, checking that no value exists with the supplied key and then puts a new value into the world state, converting the passed value to bytes as required.

The second function to add to the contract is `Update`, this will work in the same way as the `Create` function however instead of erroring if the key exists in the world state it will error if it does not.

```
// Update changes the value with key in the world state
func (s *Simple) Update(ctx *contractapi.TransactionContext, key string, value string) error {
    existing, err := ctx.GetStub().GetState(key)

    if err != nil {
        return errors.New("Unable to interact with world state")
    }

    if existing == nil {
        return fmt.Errorf("Cannot update world state pair with key %s. Does not exist", key)
    }

    err = ctx.GetStub().PutState(key, []byte(value))

    if err != nil {
        return errors.New("Unable to interact with world state")
    }

    return nil
}
```

The third and final function to add to the simple contract is `Read` this will take in a key and return the world state value. It will therefore return a string type (the value type before converting to bytes for the world state) and will also return an error type.

```
// Read returns the value at key in the world state
func (s *Simple) Read(ctx *contractapi.TransactionContext, key string, value string) (string, error) {
    existing, err := ctx.GetStub().GetState(key)

    if err != nil {
        return "", errors.New("Unable to interact with world state")
    }

    if existing == nil {
        return "", fmt.Errorf("Cannot read world state pair with key %s. Does not exist", key)
    }

    return string(existing), nil
}
```

Your final contract will then look like this:

```
package contracts

import (
    "errors"
    "fmt"

    "github.com/awjh-ibm/fabric-go-developer-api/contractapi"
)

// Simple contract for handling writing and reading from the world state
type Simple struct {
    contractapi.Contract
}

// Create adds a new key with value to the world state
func (s *Simple) Create(ctx *contractapi.TransactionContext, key string, value string) error {
    existing, err := ctx.GetStub().GetState(key)

    if err != nil {
        return errors.New("Unable to interact with world state")
    }

    if existing != nil {
        return fmt.Errorf("Cannot create world state pair with key %s. Already exists", key)
    }

    err = ctx.GetStub().PutState(key, []byte(value))

    if err != nil {
        return errors.New("Unable to interact with world state")
    }

    return nil
}

// Update changes the value with key in the world state
func (s *Simple) Update(ctx *contractapi.TransactionContext, key string, value string) error {
    existing, err := ctx.GetStub().GetState(key)

    if err != nil {
        return errors.New("Unable to interact with world state")
    }

    if existing == nil {
        return fmt.Errorf("Cannot update world state pair with key %s. Does not exist", key)
    }

    err = ctx.GetStub().PutState(key, []byte(value))

    if err != nil {
        return errors.New("Unable to interact with world state")
    }

    return nil
}

// Read returns the value at key in the world state
func (s *Simple) Read(ctx *contractapi.TransactionContext, key string, value string) (string, error) {
    existing, err := ctx.GetStub().GetState(key)

    if err != nil {
        return "", errors.New("Unable to interact with world state")
    }

    if existing == nil {
        return "", fmt.Errorf("Cannot read world state pair with key %s. Does not exist", key)
    }

    return string(existing), nil
}
```

## Using contracts in chaincode
In `go-developer-api-tutorial` create a new file called `main.go`, within this import the contractapi and your contracts package and add a main function:

```
import (
    "errors"
    "fmt"
    "contracts"

    "github.com/awjh-ibm/fabric-go-developer-api/contractapi"
)

func main() {
}
```

Within the main function create a new instance of the `Simple` struct from contracts and then use this as an argument for creating a new chaincode:

```
    simpleContract := new(contract.Simple)

    if err := contractapi.CreateNewChaincode(simpleContract); err != nil {
        fmt.Printf("Error starting simple contract chaincode: %s", err)
    }
```

Your main.go file should now look like this:

```
import (
    "fmt"
    "contracts"

    "github.com/awjh-ibm/fabric-go-developer-api/contractapi"
)

func main() {
    simpleContract := new(contracts.Simple)

    if err := contractapi.CreateNewChaincode(simpleContract); err != nil {
        fmt.Printf("Error starting simple contract chaincode: %s", err)
    }
}
```

## Running your chaincode as a developer
Open a terminal at `$GOPATH/src/github.com/hyperledger/fabric-samples/chaincode-docker-devmode`. In here run the following command to startup a simple development fabric network and launch two further containers:

```
docker-compose -f docker-compose-simple.yaml up
```

In a new terminal window enter the chaincode docker container bash environment:

```
docker exec -it chaincode bash
```

Move within that docker environment to your chaincode folder and build your chaincode:

```
cd go-developer-api-tutorial
go build
```

Now run the chaincode (Note: it should not exit):

```
CORE_PEER_ADDRESS=peer:7052 CORE_CHAINCODE_ID_NAME=mycc:0 ./go-developer-api-tutorial
```

## Interacting with your running chaincode
In a new terminal window enter the cli docker container:

```
docker exec -it cli bash
```

From here you can install, instantiate and interact with your chaincode. Start by installing the chaincode:

```
peer chaincode install -p chaincodedev/chaincode/go-developer-api-tutorial -n mycc -v 0
```

Next instantiate:

```
peer chaincode instantiate -n mycc -v 0 -c '{"Args":[]}' -C myc
```

Passing no arguments to instantiate means that no function of your contract is called and the chaincode returns shim.Success (provided that nothing else is wrong e.g. chaincode name).

Now that the chaincode is instantiated you can interact with your chaincode using invoke and query. First use an invoke to create a new key pair in the world state:

```
peer chaincode invoke -n mycc -c '{"Args":["Create", "KEY_1", "VALUE_1"]}' -C myc
``

The first argument of invoke is the function you wish to call, in this case "Create". The remaining arguments are the values to be passed to the function so in the case of create the key and value parameters. Note that you do not pass the transaction context, this is generated by the running chaincode to be passed to your contract function.

Once your key value pair is created you can use the "Update" function of your contract to change the value. This again can be done using an invoke argument:

```
peer chaincode invoke -n mycc -c '{"Args":["Update", "KEY_1", "VALUE_2"]}' -C myc
```

If you wish to read the value stored for a particular key you can query the "Read" function of the contract:

```
peer chaincode query -n mycc -c '{"Args":["Read", "KEY_1"]}' -C myc
```