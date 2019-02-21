# Incorporating multiple contracts and handling objects

## Tutorial contents
- [Prerequisites](#prerequisites)
- [Building a contract to handle an asset](#building-a-contract-to-handle-an-asset)
- [Adding a second contract to the chaincode](#adding-a-second-contract-to-the-chaincode)
- [Using a custom name for your contracts](#using-a-custom-name-for-your-contracts)
- [Interacting with your running chaincode](#interacting-with-your-running-chaincode)

The previous two tutorials built up a simple contract using the contract API to be executed using chaincode. This tutorial will cover building a more complex contract and adding both the simple and complex contracts to a single chaincode.

## Prerequisites
- Completion of ["Extending a simple contract"](./a_simple_contract) ([get the code](./tutorials/code/simple_contract_extended))

## Building a contract to handle an asset
The contract that will be written in this tutorial will manage an asset rather than just reading and writing simple values to the world state. The definition of an asset is up to the developer and is not part of the contract API therefore this tutorial will define a basic asset that is designed only for use in this tutorial. Create a new folder called defs inside the vendor folder and inside that a file called `asset.go`. Add to that file a definition of a basic asset and use JSON tags in the struct definition, this will allow the contract to store the asset as a string in the world state and also enable the contractapi to handle conversion to and from a string for the asset.

```
package defs

// BasicAsset a basic asset
type BasicAsset struct {
    ID string `json:"id"`
    Owner string `json:"owner"`
    Value int `json:"value"`
    Condition int `json:"condition"`
}

// SetConditionNew set the condition of the asset to mark as new
func(ba *BasicAsset) SetConditionNew() {
    ba.Condition = 0
}

// SetConditionUsed set the condition of the asset to mark as used
func(ba *BasicAsset) SetConditionUsed() {
    ba.Condition = 1
}
```

Note that although the asset has functions these are not the business logic. Business logic should not be placed in the asset but instead in the contract.

Now that the asset is defined, create a new contract to handle the asset. This contract will handle the business logic of managing that asset. Create the contract in the same way as the simple contract by first creating a new file inside vendor/contracts. Call this file complex.go. Inside this create a struct called Complex which embeds the `contractapi.Contract` struct to ensure it meets the `contractapi.ContractInterface`.

```
package contracts

import "github.com/awjh-ibm/fabric-go-developer-api/contractapi"

// Complex contract for handling the business logic of a basic asset
type Complex struct {
	contractapi.Contract
}
```

Now add the first function for the managing of the asset, this will be a function to create a new instance of the asset and record this in the world state using the ID as the key. The function, as will all others in this contract, will perform a get of the passed ID from the world state therefore it makes sense to use the same process as in the simple contract, a before function which calls get using the passed ID. This before function will be the same as used by the simple contract although we could have written a custom function just for this contract. Like with the simple contract it is necessary to alert the contract API to the use of this function but that will be set later. As the get function uses a custom context (utils.CustomTransactionContext) the new asset function (and all other in this contract) will use the same transaction context type.

```
// NewAsset adds a new basic asset to the world state using id as key
func (s *Complex) NewAsset(ctx *utils.CustomTransactionContext, id string, owner string, value int) error {
	existing := ctx.CallData

	if existing != nil {
		return fmt.Errorf("Cannot create new asset in world state as key %s already exists", id)
	}

	ba := defs.BasicAsset{}
	ba.ID = id
	ba.Owner = owner
	ba.Value = value
	ba.SetConditionNew()

	baBytes, _ := json.Marshal(ba)

	err := ctx.GetStub().PutState(id, []byte(baBytes))

	if err != nil {
		return errors.New("Unable to interact with world state")
	}

	return nil
}
```

The next functions will handle updating the asset in the world state. The first function will update the owner by simply replacing the owner value and the second will update the value by adding the value passed. The change of ownership will also mark the asset as used. Both functions will take the data from the world state and convert it back to a BasicAsset before updating the values.

```
// UpdateOwner changes the ownership of a basic asset and mark it as used
func (c *Complex) UpdateOwner(ctx *utils.CustomTransactionContext, id string, newOwner string) error {
	existing := ctx.CallData

	if existing == nil {
		return fmt.Errorf("Cannot update asset in world state as key %s does not exist", id)
	}

	ba := defs.BasicAsset{}

	err := json.Unmarshal(existing, &ba)

	if err != nil {
		return fmt.Errorf("Data retrieved from world state for key %s was not of type BasicAsset", id)
	}

	ba.Owner = newOwner
	ba.SetConditionUsed()

	baBytes, _ := json.Marshal(ba)

	err = ctx.GetStub().PutState(id, []byte(baBytes))

	if err != nil {
		return errors.New("Unable to interact with world state")
	}

	return nil
}

// UpdateValue changes the value of a basic asset to add the value passed
func (c *Complex) UpdateValue(ctx *utils.CustomTransactionContext, id string, valueAdd int) error {
	existing := ctx.CallData

	if existing == nil {
		return fmt.Errorf("Cannot update asset in world state as key %s does not exist", id)
	}

	ba := defs.BasicAsset{}

	err := json.Unmarshal(existing, &ba)

	if err != nil {
		return fmt.Errorf("Data retrieved from world state for key %s was not of type BasicAsset", id)
	}

	ba.Value += valueAdd

	baBytes, _ := json.Marshal(ba)

	err = ctx.GetStub().PutState(id, []byte(baBytes))

	if err != nil {
		return errors.New("Unable to interact with world state")
	}

	return nil
}
```

Add a final function to return the asset from the world state to the user.

```
// GetAsset returns the asset with id given from the world state
func (c *Complex) GetAsset(ctx *utils.CustomTransactionContext, id string) (*defs.BasicAsset, error) {
	existing := ctx.CallData

	if existing == nil {
		return nil, fmt.Errorf("Cannot read world state pair with key %s. Does not exist", id)
	}

	ba := new(defs.BasicAsset)

	err := json.Unmarshal(existing, ba)

	if err != nil {
		return nil, fmt.Errorf("Data retrieved from world state for key %s was not of type BasicAsset", id)
	}

	return ba, nil
}
```

`GetAsset` returns a pointer to a BasicAsset. The contractapi will handle conversion of this to a stringified JSON format. As this process uses the built in go JSON marshalling and unmarshalling process it is possible to customise this process by defining your own Marshal and Unmarshal functions for the asset.

You may notice that each of the functions that reads from the world state checks that the data returned is that of a basic asset and therefore this could be parted out to a separate function. As that is outside the scope of the contract API it has been ignored by this tutorial.

## Adding a second contract to the chaincode
Your main.go file will already contain the code to use the simple contract inside chaincode, here you now need to add code to use your complex contract inside the chaincode as well. This is done using the exact same method as the Simple contracy by creating a new instance of the `Complex` struct from contracts and pass this new instance as an argument to the `contractapi.CreateNewChaincode`.

```
func main() {
	simpleContract := new(contracts.Simple)
	simpleContract.SetTransactionContextHandler(new(utils.CustomTransactionContext))
	simpleContract.SetBeforeTransaction(utils.GetWorldState)
	simpleContract.SetUnknownTransaction(utils.UnknownTransactionHandler)

	complexContract := new(contracts.Complex)

	cc := contractapi.CreateNewChaincode(simpleContract, complexContract)

	if err := cc.Start(); err != nil {
		fmt.Printf("Error starting multiple contract chaincode: %s", err)
	}
}
```

Earlier in this tutorial you coded the Complex contract to make use of `utils.CustomTransactionContext` and rely on `utils.GetWorldState` being called before each function. Update the main function to tell the contract API to use these.

```
func main() {
	simpleContract := new(contracts.Simple)
	simpleContract.SetTransactionContextHandler(new(utils.CustomTransactionContext))
	simpleContract.SetBeforeTransaction(utils.GetWorldState)
	simpleContract.SetUnknownTransaction(utils.UnknownTransactionHandler)

	complexContract := new(contracts.Complex)
	complexContract.SetTransactionContextHandler(new(utils.CustomTransactionContext))
	complexContract.SetBeforeTransaction(utils.GetWorldState)

	cc := contractapi.CreateNewChaincode(simpleContract, complexContract)

	if err := cc.Start(); err != nil {
		fmt.Printf("Error starting multiple contract chaincode: %s", err)
	}
}

```

## Using a custom name for your contracts

The contract API calls the function `GetName` of the contract to decide what name to give the contract in the chaincode. If the value returned is a blank string then the contract API uses the name of the struct. By default contractapi.Contract.GetName returns a blank string and therefore the contracts which embed contractapi.Contract, as those in this tutorial do, by default have the Struct name used for their name. To use a custom name it is therefore necessary to have the `GetName` function of the contract return a non blank string. This can be done for contracts embedding `contractapi.Contract` by calling `SetName` on the contract before the chaincode is created. Update your main function to use custom names:

```
func main() {
	simpleContract := new(contracts.Simple)
	simpleContract.SetTransactionContextHandler(new(utils.CustomTransactionContext))
	simpleContract.SetBeforeTransaction(utils.GetWorldState)
	simpleContract.SetUnknownTransaction(utils.UnknownTransactionHandler)
	simpleContract.SetName("SimpleContract")

	complexContract := new(contracts.Complex)
	complexContract.SetTransactionContextHandler(new(utils.CustomTransactionContext))
	complexContract.SetBeforeTransaction(utils.GetWorldState)

	complexContract.SetName("ComplexContract")

	cc := contractapi.CreateNewChaincode(simpleContract, complexContract)

	if err := cc.Start(); err != nil {
		fmt.Printf("Error starting multiple contract chaincode: %s", err)
	}
}
```

The simple and complex contracts are now named `SimpleContract` and `ComplexContract` respectively. 

## Interacting with your running chaincode
Before starting your chaincode take down the previous tutorials docker setup using: 

```
docker-compose -f docker-compose-simple.yaml down --volume
```

The chaincode can be started in the same way as the simple contract chaincode. These instructions can be viewed [here](./a_simple_contract.md#running-your-chaincode-as-a-developer). Once you have your chaincode running you can interact with both contracts using the docker cli container. Enter this in a new terminal window:

```
docker exec -it cli bash
```

From here you can install, instantiate and interact with your chaincode. Start by installing the chaincode:

```
peer chaincode install -p chaincodedev/chaincode/go-developer-api-tutorial -n mycc -v 0
```

Next instantiate the chaincode. You can only call instantiate once on a chaincode, you do not call it for each contract. If you need to call an instantiate on both contracts it is recommended that you write a callable function of your chaincode that will call both these functions. As neither of your contracts have any functions that need to be called on instantiation you should not pass a function to be called to the instantiation request. 

```
peer chaincode instantiate -n mycc -v 0 -c '{"Args":[]}' -C myc
```

Now that the chaincode is instantiated you can invoke and query contracts within the chaincode, since you have set a name for both contracts each function call must be prefixed with the name in the format SET_CONTRACT_NAME:FUNCTION. The exception to this is that as the simple contract was passed first when creating chaincode, its functions can be called without passing the name. Run the following commands to use your simple contract:

```
peer chaincode invoke -n mycc -c '{"Args":["SimpleContract:Create", "KEY_1", "VALUE_1"]}' -C myc

peer chaincode invoke -n mycc -c '{"Args":["SimpleContract:Update", "KEY_1", "VALUE_2"]}' -C myc

peer chaincode query -n mycc -c '{"Args":["SimpleContract:Read", "KEY_1"]}' -C myc
```

You can interact with your complex contract using these commands:

```
peer chaincode invoke -n mycc -c '{"Args":["ComplexContract:NewAsset", "ASSET_1", "OWNER_1", "100"]}' -C myc

peer chaincode invoke -n mycc -c '{"Args":["ComplexContract:UpdateOwner", "ASSET_1", "OWNER_2"]}' -C myc

peer chaincode invoke -n mycc -c '{"Args":["ComplexContract:UpdateValue", "ASSET_1", "100"]}' -C myc

peer chaincode query -n mycc -c '{"Args":["ComplexContract:GetAsset", "ASSET_1"]}' -C myc
```

Note that in the complex contract it is specified that the value to be passed in as a value in both NewAsset and UpdateValue is an int yet in the commands they are strings. Fabric expects all arguments to be strings and therefore the data must be passed as a string through the above commands. The contractapi then turns that data to the expected type for the functions (assumes base 10 for numeric types). The contract API will return an error response to the peer if this conversion fails.

You can then query the system contract to see the information about both your contracts in the chaincode:

```
peer chaincode query -n mycc -c '{"Args":["org.hyperledger.fabric:GetMetadata"]}' -C myc
```