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
	"encoding/json"
	"errors"
	"fmt"

	"github.com/awjh-ibm/fabric-go-developer-api/contractapi"
)

// TransactionContext custom transaction context so can store values
type TransactionContext struct {
	contractapi.TransactionContext
	data []byte
}

// SimpleAsset - a simple asset to be managed as a key val pair
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

// ComplexAsset - a complex asset to be managed as a key val pair using JSON
type ComplexAsset struct {
	contractapi.Contract
	Owner   string   `json:"owner"`
	Colours []string `json:"colours"`
	Value   float64  `json:"value"`
}

// Create - Initialises a complex asset with the given ID in the world state
func (ca *ComplexAsset) Create(ctx *TransactionContext, assetID string) error {
	existing := ctx.data

	if len(existing) > 0 {
		return fmt.Errorf("Cannot create asset. Asset with id %s already exists", assetID)
	}

	ca.Owner = "Regulator"
	ca.Value = 0
	ca.Colours = []string{}

	return ca.put(ctx, assetID)
}

// UpdateOwner - Updates a complex asset with given ID in the world state to have a new owner
func (ca *ComplexAsset) UpdateOwner(ctx *TransactionContext, assetID string, newOwner string) error {
	existing := ctx.data

	if len(existing) == 0 {
		return fmt.Errorf("Cannot update asset. Asset with id %s does not exist", assetID)
	}

	err := json.Unmarshal(existing, ca)

	if err != nil {
		return fmt.Errorf("Asset with id %s is not a ComplexAsset", assetID)
	}

	ca.Owner = newOwner

	return ca.put(ctx, assetID)
}

// UpdateValue - Updates a complex asset with given ID in the world state to have a new value by adding the passed value to its existing value
func (ca *ComplexAsset) UpdateValue(ctx *TransactionContext, assetID string, additionalValue float64) error {
	existing := ctx.data

	if len(existing) == 0 {
		return fmt.Errorf("Cannot update asset. Asset with id %s does not exist", assetID)
	}

	err := json.Unmarshal(existing, ca)

	if err != nil {
		return fmt.Errorf("Asset with id %s is not a ComplexAsset", assetID)
	}

	ca.Value += additionalValue

	return ca.put(ctx, assetID)
}

// AddColours - add an array of new colours to existing list
func (ca *ComplexAsset) AddColours(ctx *TransactionContext, assetID string, additionalColours []string) error {
	existing := ctx.data

	if len(existing) == 0 {
		return fmt.Errorf("Cannot update asset. Asset with id %s does not exist", assetID)
	}

	err := json.Unmarshal(existing, ca)

	if err != nil {
		return fmt.Errorf("Asset with id %s is not a ComplexAsset", assetID)
	}

	ca.Colours = append(ca.Colours, additionalColours...)

	return ca.put(ctx, assetID)
}

// Read - Returns the complex asset with given ID from world state as string
func (ca *ComplexAsset) Read(ctx *TransactionContext, assetID string) (string, error) {
	existing := ctx.data

	if len(existing) == 0 {
		return "", fmt.Errorf("Cannot read asset. Asset with id %s does not exist", assetID)
	}

	err := json.Unmarshal(existing, ca)

	if err != nil {
		return "", fmt.Errorf("Asset with id %s is not a ComplexAsset", assetID)
	}

	return ca.Owner + " - " + fmt.Sprint(ca.Value) + " - " + fmt.Sprint(ca.Colours), nil
}

// ReadValue - Returns the value of a complex asset with given ID from world state
func (ca *ComplexAsset) ReadValue(ctx *TransactionContext, assetID string) (float64, error) {
	existing := ctx.data

	if len(existing) == 0 {
		return -1, fmt.Errorf("Cannot read asset. Asset with id %s does not exist", assetID)
	}

	err := json.Unmarshal(existing, ca)

	if err != nil {
		return -1, fmt.Errorf("Asset with id %s is not a ComplexAsset", assetID)
	}

	return ca.Value, nil
}

// ReadColours - Returns the colours of a complex asset with given ID from world state
func (ca *ComplexAsset) ReadColours(ctx *TransactionContext, assetID string) ([]string, error) {
	existing := ctx.data

	if len(existing) == 0 {
		return nil, fmt.Errorf("Cannot read asset. Asset with id %s does not exist", assetID)
	}

	err := json.Unmarshal(existing, ca)

	if err != nil {
		return nil, fmt.Errorf("Asset with id %s is not a ComplexAsset", assetID)
	}

	return ca.Colours, nil
}

func (ca *ComplexAsset) put(ctx *TransactionContext, assetID string) error {
	caJSON, err := json.Marshal(&ca)

	if err != nil {
		return errors.New("Error converting asset to JSON")
	}

	err = ctx.GetStub().PutState(assetID, caJSON)

	if err != nil {
		return errors.New("Unable to interact with world state")
	}

	return nil
}

func getAsset(ctx *TransactionContext, assetID string) error {

	existing, err := ctx.GetStub().GetState(assetID)

	if err != nil {
		return errors.New("Unable to interact with world state")
	}

	ctx.data = existing

	return nil
}

func handleSimpleUnknown(ctx *TransactionContext) error {
	fn, args := ctx.GetStub().GetFunctionAndParameters()

	return fmt.Errorf("Unknown function name %s passed to simple asset with args %v", fn, args)
}

func handleComplexUnknown(ctx *TransactionContext) error {
	fn, args := ctx.GetStub().GetFunctionAndParameters()

	return fmt.Errorf("Unknown function name %s passed to complex asset with args %v", fn, args)
}

func main() {
	sac := new(SimpleAsset)
	sac.SetTransactionContextHandler(new(TransactionContext))
	sac.SetBeforeTransaction(getAsset)
	sac.SetUnknownTransaction(handleSimpleUnknown)
	sac.SetName("simpleasset")

	cac := new(ComplexAsset)
	cac.SetTransactionContextHandler(new(TransactionContext))
	cac.SetBeforeTransaction(getAsset)
	cac.SetUnknownTransaction(handleComplexUnknown)
	cac.SetName("complexasset")

	if err := contractapi.CreateNewChaincode(sac, cac); err != nil {
		fmt.Printf("Error starting multi asset chaincode: %s", err)
	}
}
