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

package utils

import (
	"errors"
	"fmt"

	"github.com/awjh-ibm/fabric-go-developer-api/contractapi"
	"github.com/hyperledger/fabric/core/chaincode/shim"
)

// CustomTransactionContext adds extra field to contractapi.TransactionContext
// so that data can be between calls
type CustomTransactionContext struct {
	contractapi.TransactionContext
	CallData []byte
}

// GetWorldState takes a key and sets what is found in the world state for that
// key in the transaction context
func GetWorldState(ctx *CustomTransactionContext) error {
	_, params := ctx.GetStub().GetFunctionAndParameters()

	if len(params) < 1 {
		return errors.New("Missing key for world state")
	}

	existing, err := ctx.GetStub().GetState(params[0])

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
