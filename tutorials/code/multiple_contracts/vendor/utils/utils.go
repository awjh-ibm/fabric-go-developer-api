package utils

import (
	"errors"
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"

	"github.com/awjh-ibm/fabric-go-developer-api/contractapi"
)

// CustomTransactionContext adds extra field to contractapi.TransactionContext
// so that data can be between calls
type CustomTransactionContext struct {
	contractapi.TransactionContext
	CallData []byte
}

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

var logger = shim.NewLogger("go-developer-api-tutorial")

// UnknownTransactionHandler logs details of a bad transaction request
// and returns a shim error
func UnknownTransactionHandler(ctx *CustomTransactionContext) error {
	fcn, args := ctx.GetStub().GetFunctionAndParameters()
	logger.Errorf("Invalid function %s passed with args %v", fcn, args)
	return fmt.Errorf("Invalid function %s", fcn)
}
