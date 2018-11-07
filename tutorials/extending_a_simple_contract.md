# Extending a simple contract
- [Prerequisites](#prerequisites)
- [Calling functions every time a request is made and using custom transaction contexts](#calling-functions-every-time-a-request-is-made-and-using-custom-transaction-contexts)
- [Handling unknown function requests](#handling-unknown-function-requests)
- [Interacting with your running chaincode](#interacting-with-your-running-chaincode)

The tutorial ["A simple contract"](./a_simple_contract) covered the basics of using the contract API. This tutorial will build on top of the code created there to use some more features of the contract API.

## Prerequisites
- Completion of ["A simple contract"](./a_simple_contract) ([get the code](./tutorials/code/simple_contract))

## Calling functions every time a request is made and using custom transaction contexts
The contract API provides functionality for specified functions to be called before and after each call to a contract. The simple contract made in the previous tutorial performs the same task at the start of each function call, getting data from the world state. It would therefore be useful to create one function to do this and set it up to run before each transaction. The transaction context is the same for all function calls during a trasnsaction so data set in the transaction context by the before call can be used in the main and after calls. Likewise data set/updated by the main call can be used in the after call. Updates to the world state made by previous functions are not readable by later functions as all the calls are made within the same transaction and in fabric you [cannot read your own writes](https://hyperledger-fabric.readthedocs.io/en/master/readwrite.html).

Before and after functions must follow the same structure rules as contract functions for parameters and returns, however they do not have to be public or attached to a contract struct. As the transaction context is the same for all calls of a transaction the parameters sent in will be the same for the before, main and after functions. It is important to note that this does not mean that the parameter structure has to be the same for each as the contract API will only send as many parameters as the function requires but the order must match. It is also possible to access the sent parameters in their entirety (and the function call) via the [stub](https://godoc.org/github.com/hyperledger/fabric/core/chaincode/shim#ChaincodeStub).

If a before transaction function is defined to return an error and returns a non nil error value when called the remaining function calls are not made and a shim.Error is returned to the peer with the before transaction function's returned error value. Likewise if the main function errors the after function is not called and the error is returned to the peer. If an after transaction function errors then again the shim recieves that error. If the after transaction function does not return an error then the success response from the main function is returned to the peer whether the after function has a success response or not. 

To create your before transaction function create a new folder called utils inside vendor and create a file utils.go. In here import the contract API and specify a struct to define a custom transaction context to allow data to be passed from the before function to the after function. You could implement the contractapi.TransactionContextInterface manually but as in this case it involves only adding a new field to the context, embed the contractapi.TransactionContext inside your own.

```
package utils

import (
	"errors"

	"github.com/awjh-ibm/fabric-go-developer-api/contractapi"
)

// CustomTransactionContext adds extra field to contractapi.TransactionContext
// so that data can be between calls
type CustomTransactionContext struct {
	contractapi.TransactionContext
	CallData []byte
}
```

In `go-developer-api-tutorial/vendor/contracts/simple.go` import utils and update each function to use `*utils.CustomTransactionContext` instead of `*contractapi.TransactionContext`. The contract API must be informed to use this new transaction context when calling the Simple contract. The contract API uses `contractapi.ContractInterface.GetTransactionContextHandler()` to get the transaction context to use. Since the Simple contract embeds the `contractapi.Contract` struct we can set the value to be returned when that function is called for Simple contract by calling `SetTransactionContextHandler()`. Add this to the start of your main function inside main.go, passing the new transaction context. Ensure that the setting of the transaction context occurs before the new chaincode is created.

```
simpleContract.SetTransactionContextHandler(new(utils.CustomTransactionContext))
```

Now in `vendor/utils/utils.go` add the function to get the state details. The function should be public so that it can be accessed by main and take in the custom transaction context and a key string param. As the function is only going to be used as a before transaction function there is no need to have it return any type other than error.

```
// GetWorldState takes a key and sets what is found in the world state for that
// key in the transaction context
func GetWorldState(ctx *CustomTransactionContext, key string) error {
	existing, err := ctx.GetStub().GetState(key)

	if err != nil {
		return errors.New("Unable to interact with world state")
	}

	ctx.CallData = existing

	return nil
}

```

The function uses the key passed in the param (note that it is the first parameter in all the contract functions as each access the same args list) to get data from the world state. It the updates the transaction context to store this data so that later function can use it.

Like in the case of the custom transaction context, the contract API needs to be informed of the changes. The contractapi uses `contractapi.ContractInterface.GetBeforeTransaction()` to get which (if any) transaction it should call before the main transaction. As the Simple contract embeds the `contractapi.Contract` struct we can set the value to be returned when that is called for Simple contract by calling `SetBeforeTransaction()`. This needs to be added inside the main function of main.go before the new chaincode is created. Your main.go file should now look like this:

```
package main

import (
	"contracts"
	"fmt"
	"utils"

	"github.com/awjh-ibm/fabric-go-developer-api/contractapi"
)

func main() {
	simpleContract := new(contracts.Simple)
	simpleContract.SetTransactionContextHandler(new(utils.CustomTransactionContext))
	simpleContract.SetBeforeTransaction(utils.GetWorldState)

	if err := contractapi.CreateNewChaincode(simpleContract); err != nil {
		fmt.Printf("Error starting simple contract chaincode: %s", err)
	}
}
```

As the utils.GetWorldState function is set to be called before each transaction that calls simpleContract you can now remove the code to get data from the world state from each of the contract's functions. To do this replace:

```
existing, err := ctx.GetStub().GetState(key)

if err != nil {
    return errors.New("Unable to interact with world state")
}
```

with

```
existing := ctx.CallData
```

> Note: As you have removed the definition of err you will need to change err = to err := in the Create and Update functions

## Handling unknown function requests
By default if a function is passed to an Init/Invoke that is unknown to the chaincode, for example when a user misspells a known function or enters a non-existant one, the chaincode returns an error response to the peer to let the user know of the issue. It is possible however to specify a custom handler for these unknown function for your contract. Like the before and after transactions functions the unknown transaction function must match the param and return types of a standard contract function. Again like the before and after transaction functions the unknown transaction function does not have to be public or a function of the contract, but it can be. Unknown transaction functions do NOT have to return an error, if they do then the after transaction function will not be called.

The function for this tutorial will handle logging the details of the call using the shim logger and return an error. The function will not be reading or writing anything to the world state but will still take the transaction context as it will rely on the stub for accessing the call details. As the function will be a catch all for all unknown calls the function defined here will not take any parameters, as it is not possible to know how many arguments a user may provide in error, instead it will use the stub to read the passed data. You should define this function in your utils.go file.

```
var logger = shim.NewLogger("go-developer-api-tutorial")

// UnknownTransactionHandler logs details of a bad transaction request
// and returns a shim error
func UnknownTransactionHandler(ctx *CustomTransactionContext) error {
	fcn, args := ctx.GetStub().GetFunctionAndParameters()
	logger.Errorf("Invalid function %s passed with args %v", fcn, args)
	return fmt.Errorf("Invalid function %s", fcn)
}
```

Also include fmt in your imports.

Like with the other extended settings you must inform the contract API of your intent to use a custom unknown transaction handler. The contract API finds which function to call by using the `contractapi.ContractInterface.GetUnknownTransaction()` function. As the Simple contract embeds the `contractapi.Contract` struct you can set the value returned by this function using `SetUnknownTransaction()` and passing in a reference to your function. Add the call to this function in your main.go file above the creation of the new chaincode.

```
simpleContract.SetUnknownTransaction(utils.UnknownTransactionHandler)
```

## Interacting with your running chaincode
This extended chaincode should run in the exact same way as the simple chaincode from the previous tutorial. You can therefore run the chaincode by following the previous tutorials [instructions](./a_simple_contract.md#interacting-with-your-running-chaincode).