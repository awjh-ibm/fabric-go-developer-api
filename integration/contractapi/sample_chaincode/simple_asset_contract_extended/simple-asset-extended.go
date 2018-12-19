/*
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"errors"
	"fmt"

	"github.com/awjh-ibm/fabric-go-developer-api/contractapi"
)

// TransactionContext custom transaction context so can store values
type TransactionContext struct {
	contractapi.TransactionContext
	data []byte
}

// SimpleAsset with biz logic
type SimpleAsset struct {
	contractapi.Contract
}

// Create - Initialises a simple asset with the given ID in the world state
func (sa *SimpleAsset) Create(ctx *TransactionContext, assetID string) error {
	existing := ctx.data

	if len(existing) > 0 {
		return fmt.Errorf("Cannot create asset. Asset with id %s already exists", assetID)
	}

	err := ctx.GetStub().PutState(assetID, []byte("Initialised"))

	if err != nil {
		return errors.New("Unable to interact with world state")
	}

	return nil
}

// Update - Updates a simple asset with given ID in the world state
func (sa *SimpleAsset) Update(ctx *TransactionContext, assetID string, value string) error {
	existing := ctx.data

	if len(existing) == 0 {
		return fmt.Errorf("Cannot update asset. Asset with id %s does not exist", assetID)
	}

	err := ctx.GetStub().PutState(assetID, []byte(value))

	if err != nil {
		return errors.New("Unable to interact with world state")
	}

	return nil
}

// Read - Returns value of a simple asset with given ID from world state as string
func (sa *SimpleAsset) Read(ctx *TransactionContext, assetID string) (string, error) {
	existing := ctx.data

	if len(existing) == 0 {
		return "", fmt.Errorf("Cannot read asset. Asset with id %s does not exist", assetID)
	}

	return string(string(existing)), nil
}

func getAsset(ctx *TransactionContext, assetID string) error {

	existing, err := ctx.GetStub().GetState(assetID)

	if err != nil {
		return errors.New("Unable to interact with world state")
	}

	ctx.data = existing

	return nil
}

func handleUnknown(ctx *TransactionContext) error {
	fn, args := ctx.GetStub().GetFunctionAndParameters()

	return fmt.Errorf("Unknown function name %s passed with args %v", fn, args)
}

func main() {
	sac := new(SimpleAsset)
	sac.SetTransactionContextHandler(new(TransactionContext))
	sac.SetBeforeTransaction(getAsset)
	sac.SetUnknownTransaction(handleUnknown)

	cc := contractapi.CreateNewChaincode(sac)

	if err := cc.Start(); err != nil {
		fmt.Printf("Error starting SimpleAsset chaincode: %s", err)
	}
}
