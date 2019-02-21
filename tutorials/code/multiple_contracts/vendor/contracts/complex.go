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

package contracts

import (
	"defs"
	"encoding/json"
	"errors"
	"fmt"
	"utils"

	"github.com/awjh-ibm/fabric-go-developer-api/contractapi"
)

// Complex contract for handling the business logic of a basic asset
type Complex struct {
	contractapi.Contract
}

// NewAsset adds a new basic asset to the world state using id as key
func (c *Complex) NewAsset(ctx *utils.CustomTransactionContext, id string, owner string, value int) error {
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
