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

package contractapi

type systemContract struct {
	Contract
	metadata string
}

func (sc *systemContract) setMetadata(metadata string) {
	sc.metadata = metadata
}

// GetMetadata returns JSON formatted metadata of chaincode
// the system contract is part of. This metadata is composed
// of reflected metadata combined with the metadata file
// if used
func (sc *systemContract) GetMetadata() string {
	return sc.metadata
}
