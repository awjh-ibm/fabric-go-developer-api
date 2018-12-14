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

package e2e

import (
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"syscall"
	"time"

	docker "github.com/fsouza/go-dockerclient"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/tedsuo/ifrit"

	"github.com/hyperledger/fabric/integration/nwo"
	"github.com/hyperledger/fabric/integration/nwo/commands"
)

var _ = Describe("contractapi - EndToEnd", func() {
	var (
		testDir   string
		client    *docker.Client
		network   *nwo.Network
		chaincode nwo.Chaincode
		process   ifrit.Process
	)

	BeforeEach(func() {
		var err error
		testDir, err = ioutil.TempDir("", "e2e")
		Expect(err).NotTo(HaveOccurred())

		client, err = docker.NewClientFromEnv()
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		if process != nil {
			process.Signal(syscall.SIGTERM)
			Eventually(process.Wait(), time.Minute).Should(Receive())
		}
		if network != nil {
			network.Cleanup()
		}
		os.RemoveAll(testDir)
	})

	Describe("single contract contractapi created chaincode", func() {
		BeforeEach(func() {
			network = nwo.New(nwo.BasicSolo(), testDir, client, 30000, components)
			network.GenerateConfigTree()
			network.Bootstrap()

			networkRunner := network.NetworkGroupRunner()
			process = ifrit.Invoke(networkRunner)
			Eventually(process.Ready()).Should(BeClosed())
		})

		It("can be deployed, invoked and queried with expected results", func() {
			chaincode = nwo.Chaincode{
				Name:    "mycc",
				Version: "0.0",
				Path:    "github.com/hyperledger/fabric/integration/contractapi/sample_chaincode/simple_asset_contract",
				Ctor:    `{"Args":["SimpleAsset:Create","ASSET_1"]}`,
				Policy:  `AND ('Org1MSP.member','Org2MSP.member')`,
			}

			orderer := network.Orderer("orderer")
			network.CreateAndJoinChannel(orderer, "testchannel")

			By("deploying the chaincode")
			nwo.DeployChaincode(network, "testchannel", orderer, chaincode)

			peer := network.Peer("Org1", "peer1")

			By("querying instantiated simple asset chaincode")
			RunSimpleQuery(network, orderer, peer, []string{"SimpleAsset:Read", "ASSET_1"}, "Initialised")

			By("querying instantiated simple asset chaincode using a blank name")
			RunSimpleQuery(network, orderer, peer, []string{"Read", "ASSET_1"}, "Initialised")

			By("invoking simple asset chaincode")
			RunSimpleInvoke(network, orderer, peer, []string{"SimpleAsset:Update", "ASSET_1", "Updated"})

			By("querying invoked simple asset chaincode")
			RunSimpleQuery(network, orderer, peer, []string{"SimpleAsset:Read", "ASSET_1"}, "Updated")

			By("querying a function that returns an error")
			RunSimpleBadQuery(network, orderer, peer, []string{"SimpleAsset:Read", "ASSET_2"}, "Cannot read asset. Asset with id ASSET_2 does not exist")

			By("invoking a function that returns an error")
			RunSimpleBadInvoke(network, orderer, peer, []string{"SimpleAsset:Update", "ASSET_2", "Update"})

			By("querying a function that does not exist")
			RunSimpleBadQuery(network, orderer, peer, []string{"SimpleAsset:BadFunction", "ASSET_1"}, "Function BadFunction not found in contract SimpleAsset")

			By("querying a name that does not exist")
			RunSimpleBadQuery(network, orderer, peer, []string{"badname:Read", "ASSET_1"}, "Contract not found with name badname")
		})
	})

	Describe("single name contractapi created chaincode using extended functions", func() {
		BeforeEach(func() {
			network = nwo.New(nwo.BasicSolo(), testDir, client, 30000, components)
			network.GenerateConfigTree()
			network.Bootstrap()

			networkRunner := network.NetworkGroupRunner()
			process = ifrit.Invoke(networkRunner)
			Eventually(process.Ready()).Should(BeClosed())
		})

		It("can be deployed, invoked and queried with expected results when using a before function", func() {
			chaincode = nwo.Chaincode{
				Name:    "mycc",
				Version: "0.0",
				Path:    "github.com/hyperledger/fabric/integration/contractapi/sample_chaincode/simple_asset_contract_extended",
				Ctor:    `{"Args":["SimpleAsset:Create","ASSET_1"]}`,
				Policy:  `AND ('Org1MSP.member','Org2MSP.member')`,
			}

			orderer := network.Orderer("orderer")
			network.CreateAndJoinChannel(orderer, "testchannel")

			By("deploying the chaincode")
			nwo.DeployChaincode(network, "testchannel", orderer, chaincode)

			peer := network.Peer("Org1", "peer1")

			By("querying instantiated simple asset extended chaincode")
			RunSimpleQuery(network, orderer, peer, []string{"SimpleAsset:Read", "ASSET_1"}, "Initialised")

			By("invoking simple asset extended chaincode")
			RunSimpleInvoke(network, orderer, peer, []string{"SimpleAsset:Update", "ASSET_1", "Updated"})

			By("querying initialised simple asset extended chaincode")
			RunSimpleQuery(network, orderer, peer, []string{"SimpleAsset:Read", "ASSET_1"}, "Updated")

			By("querying a function that returns an error")
			RunSimpleBadQuery(network, orderer, peer, []string{"SimpleAsset:Read", "ASSET_2"}, "Cannot read asset. Asset with id ASSET_2 does not exist")

			By("invoking a function that returns an error")
			RunSimpleBadInvoke(network, orderer, peer, []string{"SimpleAsset:Update", "ASSET_2", "Update"})
		})

		It("can be deployed and uses custom unknown function handler when bad function name passed", func() {
			chaincode = nwo.Chaincode{
				Name:    "mycc",
				Version: "0.0",
				Path:    "github.com/hyperledger/fabric/integration/contractapi/sample_chaincode/simple_asset_contract_extended",
				Ctor:    `{"Args":["SimpleAsset:Create","ASSET_1"]}`,
				Policy:  `AND ('Org1MSP.member','Org2MSP.member')`,
			}

			orderer := network.Orderer("orderer")
			network.CreateAndJoinChannel(orderer, "testchannel")

			By("deploying the chaincode")
			nwo.DeployChaincode(network, "testchannel", orderer, chaincode)

			peer := network.Peer("Org1", "peer1")

			By("querying instantiated simple asset extended chaincode with unknown function")
			RunSimpleBadQuery(network, orderer, peer, []string{"SimpleAsset:BadFunction", "ASSET_1"}, "Unknown function name SimpleAsset:BadFunction passed with args [ASSET_1]")
		})
	})

	Describe("multiple name contractapi created chaincode", func() {
		BeforeEach(func() {
			network = nwo.New(nwo.BasicSolo(), testDir, client, 30000, components)
			network.GenerateConfigTree()
			network.Bootstrap()

			networkRunner := network.NetworkGroupRunner()
			process = ifrit.Invoke(networkRunner)
			Eventually(process.Ready()).Should(BeClosed())
		})

		It("can be deployed, invoked and queried with expected results", func() {
			chaincode = nwo.Chaincode{
				Name:    "mycc",
				Version: "0.0",
				Path:    "github.com/hyperledger/fabric/integration/contractapi/sample_chaincode/multiple_asset_contract",
				Ctor:    `{"Args":["simpleasset:Create","SIMPLE_ASSET_1"]}`,
				Policy:  `AND ('Org1MSP.member','Org2MSP.member')`,
			}

			orderer := network.Orderer("orderer")
			network.CreateAndJoinChannel(orderer, "testchannel")

			By("deploying the chaincode")
			nwo.DeployChaincode(network, "testchannel", orderer, chaincode)

			peer := network.Peer("Org1", "peer1")

			By("querying simple asset in the chaincode")
			RunSimpleQuery(network, orderer, peer, []string{"simpleasset:Read", "SIMPLE_ASSET_1"}, "Initialised")

			By("invoking simple asset in the chaincode")
			RunSimpleInvoke(network, orderer, peer, []string{"simpleasset:Update", "SIMPLE_ASSET_1", "Updated"})

			By("querying invoked simple asset in the chaincode")
			RunSimpleQuery(network, orderer, peer, []string{"simpleasset:Read", "SIMPLE_ASSET_1"}, "Updated")

			By("invoking complex asset in the chaincode using multiple types")
			RunSimpleInvoke(network, orderer, peer, []string{"complexasset:Create", "COMPLEX_ASSET_1"})

			By("invoking complex asset in the chaincode using UpdateValue")
			RunSimpleInvoke(network, orderer, peer, []string{"complexasset:UpdateValue", "COMPLEX_ASSET_1", "101.23"})

			By("invoking complex asset in the chaincode using AddColours")
			RunSimpleInvoke(network, orderer, peer, []string{"complexasset:AddColours", "COMPLEX_ASSET_1", "[\\\"red\\\", \\\"white\\\", \\\"blue\\\"]"})

			By("querying complex asset in the chaincode")
			RunSimpleQuery(network, orderer, peer, []string{"complexasset:Read", "COMPLEX_ASSET_1"}, "Regulator - 101.23 - [red white blue]")

			By("querying a non string value of a complex asset in the chaincode")
			RunSimpleQuery(network, orderer, peer, []string{"complexasset:ReadValue", "COMPLEX_ASSET_1"}, "101.23")

			By("querying a slice value of a complex asset in the chaincode")
			RunSimpleQuery(network, orderer, peer, []string{"complexasset:ReadColours", "COMPLEX_ASSET_1"}, "[\"red\",\"white\",\"blue\"]")

			By("querying a simple asset function that returns an error")
			RunSimpleBadQuery(network, orderer, peer, []string{"simpleasset:Read", "SIMPLE_ASSET_2"}, "Cannot read asset. Asset with id SIMPLE_ASSET_2 does not exist")

			By("invoking a simple asset function that returns an error")
			RunSimpleBadInvoke(network, orderer, peer, []string{"simpleasset:Update", "SIMPLE_ASSET_2", "Update"})

			By("querying a complex asset function that returns an error")
			RunSimpleBadQuery(network, orderer, peer, []string{"complexasset:Read", "SIMPLE_ASSET_1"}, "Asset with id SIMPLE_ASSET_1 is not a ComplexAsset")

			By("invoking a complex asset function that returns an error")
			RunSimpleBadInvoke(network, orderer, peer, []string{"complexasset:UpdateOwner", "SIMPLE_ASSET_1", "Andy"})
		})

		It("can handle custom unknown functions for multiple contracts", func() {
			chaincode = nwo.Chaincode{
				Name:    "mycc",
				Version: "0.0",
				Path:    "github.com/hyperledger/fabric/integration/contractapi/sample_chaincode/multiple_asset_contract",
				Ctor:    `{"Args":["simpleasset:Create","SIMPLE_ASSET_1"]}`,
				Policy:  `AND ('Org1MSP.member','Org2MSP.member')`,
			}

			orderer := network.Orderer("orderer")
			network.CreateAndJoinChannel(orderer, "testchannel")

			By("deploying the chaincode")
			nwo.DeployChaincode(network, "testchannel", orderer, chaincode)

			peer := network.Peer("Org1", "peer1")

			By("querying instantited chaincode simpleasset name with unknown function")
			RunSimpleBadQuery(network, orderer, peer, []string{"simpleasset:BadFunction", "SIMPLE_ASSET_1"}, "Unknown function name simpleasset:BadFunction passed to simple asset with args [SIMPLE_ASSET_1]")

			By("querying instantited chaincode complexasset name with unknown function")
			RunSimpleBadQuery(network, orderer, peer, []string{"complexasset:BadFunction", "COMPLEX_ASSET_1"}, "Unknown function name complexasset:BadFunction passed to complex asset with args [COMPLEX_ASSET_1]")

			By("querying a function from another name")
			RunSimpleBadQuery(network, orderer, peer, []string{"complexasset:Update", "SIMPLE_ASSET_1"}, "Unknown function name complexasset:Update passed to complex asset with args [SIMPLE_ASSET_1]")

			By("querying using the default namespace for the non default contract")
			RunSimpleBadQuery(network, orderer, peer, []string{"ReadColours", "COMPLEX_ASSET_1"}, "Unknown function name ReadColours passed to simple asset with args [COMPLEX_ASSET_1]")
		})
	})

	Describe("simple contractapi created chaincode using contract not using contractapi.Contract", func() {
		BeforeEach(func() {
			network = nwo.New(nwo.BasicSolo(), testDir, client, 30000, components)
			network.GenerateConfigTree()
			network.Bootstrap()

			networkRunner := network.NetworkGroupRunner()
			process = ifrit.Invoke(networkRunner)
			Eventually(process.Ready()).Should(BeClosed())
		})

		It("can be deployed, invoked and queried with expected results", func() {
			chaincode = nwo.Chaincode{
				Name:    "mycc",
				Version: "0.0",
				Path:    "github.com/hyperledger/fabric/integration/contractapi/sample_chaincode/contract_interface_chaincode",
				Ctor:    `{"Args":["org.asset.simple:Create","SIMPLE_ASSET_1"]}`,
				Policy:  `AND ('Org1MSP.member','Org2MSP.member')`,
			}

			orderer := network.Orderer("orderer")
			network.CreateAndJoinChannel(orderer, "testchannel")

			By("deploying the chaincode")
			nwo.DeployChaincode(network, "testchannel", orderer, chaincode)

			peer := network.Peer("Org1", "peer1")

			By("querying simple asset in the chaincode")
			RunSimpleQuery(network, orderer, peer, []string{"org.asset.simple:Read", "SIMPLE_ASSET_1"}, "Initialised")

			By("invoking simple asset in the chaincode")
			RunSimpleInvoke(network, orderer, peer, []string{"org.asset.simple:Update", "SIMPLE_ASSET_1", "Updated"})

			By("querying initialised simple asset extended chaincode")
			RunSimpleQuery(network, orderer, peer, []string{"org.asset.simple:Read", "SIMPLE_ASSET_1"}, "Updated")

			By("querying a function that returns an error")
			RunSimpleBadQuery(network, orderer, peer, []string{"org.asset.simple:Read", "SIMPLE_ASSET_2"}, "Cannot read asset. Asset with id SIMPLE_ASSET_2 does not exist")

			By("invoking a function that returns an error")
			RunSimpleBadInvoke(network, orderer, peer, []string{"org.asset.simple:Update", "SIMPLE_ASSET_2", "Update"})
		})

		It("can be deployed and uses custom unknown function handler when bad function name passed", func() {
			chaincode = nwo.Chaincode{
				Name:    "mycc",
				Version: "0.0",
				Path:    "github.com/hyperledger/fabric/integration/contractapi/sample_chaincode/contract_interface_chaincode",
				Ctor:    `{"Args":["org.asset.simple:Create","SIMPLE_ASSET_1"]}`,
				Policy:  `AND ('Org1MSP.member','Org2MSP.member')`,
			}

			orderer := network.Orderer("orderer")
			network.CreateAndJoinChannel(orderer, "testchannel")

			By("deploying the chaincode")
			nwo.DeployChaincode(network, "testchannel", orderer, chaincode)

			peer := network.Peer("Org1", "peer1")

			By("querying instantiated simple asset extended chaincode with unknown function")
			RunSimpleBadQuery(network, orderer, peer, []string{"org.asset.simple:BadFunction", "SIMPLE_ASSET_1"}, "Unknown function name org.asset.simple:BadFunction passed with args [SIMPLE_ASSET_1]")
		})
	})

	Describe("simple contractapi created chaincode using transaction context not using contractapi.TransactionContext", func() {
		BeforeEach(func() {
			network = nwo.New(nwo.BasicSolo(), testDir, client, 30000, components)
			network.GenerateConfigTree()
			network.Bootstrap()

			networkRunner := network.NetworkGroupRunner()
			process = ifrit.Invoke(networkRunner)
			Eventually(process.Ready()).Should(BeClosed())
		})

		It("can be deployed, invoked and queried with expected results", func() {
			chaincode = nwo.Chaincode{
				Name:    "mycc",
				Version: "0.0",
				Path:    "github.com/hyperledger/fabric/integration/contractapi/sample_chaincode/transaction_context_interface_chaincode",
				Ctor:    `{"Args":["SimpleAsset:Create","SIMPLE_ASSET_1"]}`,
				Policy:  `AND ('Org1MSP.member','Org2MSP.member')`,
			}

			orderer := network.Orderer("orderer")
			network.CreateAndJoinChannel(orderer, "testchannel")

			By("deploying the chaincode")
			nwo.DeployChaincode(network, "testchannel", orderer, chaincode)

			peer := network.Peer("Org1", "peer1")

			By("querying simple asset in the chaincode")
			RunSimpleQuery(network, orderer, peer, []string{"SimpleAsset:Read", "SIMPLE_ASSET_1"}, "Initialised")

			By("invoking simple asset in the chaincode")
			RunSimpleInvoke(network, orderer, peer, []string{"SimpleAsset:Update", "SIMPLE_ASSET_1", "Updated"})

			By("querying initialised simple asset transaction context chaincode")
			RunSimpleQuery(network, orderer, peer, []string{"SimpleAsset:Read", "SIMPLE_ASSET_1"}, "Updated")

			By("querying a function that returns an error")
			RunSimpleBadQuery(network, orderer, peer, []string{"SimpleAsset:Read", "SIMPLE_ASSET_2"}, "Cannot read asset. Asset with id SIMPLE_ASSET_2 does not exist")

			By("invoking a function that returns an error")
			RunSimpleBadInvoke(network, orderer, peer, []string{"SimpleAsset:Update", "SIMPLE_ASSET_2", "Update"})
		})

		It("can be deployed and uses custom unknown function handler when bad function name passed", func() {
			chaincode = nwo.Chaincode{
				Name:    "mycc",
				Version: "0.0",
				Path:    "github.com/hyperledger/fabric/integration/contractapi/sample_chaincode/transaction_context_interface_chaincode",
				Ctor:    `{"Args":["SimpleAsset:Create","SIMPLE_ASSET_1"]}`,
				Policy:  `AND ('Org1MSP.member','Org2MSP.member')`,
			}

			orderer := network.Orderer("orderer")
			network.CreateAndJoinChannel(orderer, "testchannel")

			By("deploying the chaincode")
			nwo.DeployChaincode(network, "testchannel", orderer, chaincode)

			peer := network.Peer("Org1", "peer1")

			By("querying instantiated simple asset extended chaincode with unknown function")
			RunSimpleBadQuery(network, orderer, peer, []string{"SimpleAsset:BadFunction", "SIMPLE_ASSET_1"}, "Unknown function name SimpleAsset:BadFunction passed with args [SIMPLE_ASSET_1]")
		})
	})
})

func RunSimpleQuery(n *nwo.Network, orderer *nwo.Orderer, peer *nwo.Peer, args []string, expectedResult string) {
	queryArgs := sliceToCLIArgs(args)

	sess, err := n.PeerUserSession(peer, "User1", commands.ChaincodeQuery{
		ChannelID: "testchannel",
		Name:      "mycc",
		Ctor:      `{"Args":[` + queryArgs + `]}`,
	})

	Expect(err).NotTo(HaveOccurred())
	Eventually(sess, time.Minute).Should(gexec.Exit(0))
	Expect(sess).To(gbytes.Say(regexp.QuoteMeta(expectedResult)))
}

func RunSimpleBadQuery(n *nwo.Network, orderer *nwo.Orderer, peer *nwo.Peer, args []string, expectedResult string) {
	queryArgs := sliceToCLIArgs(args)

	sess, err := n.PeerUserSession(peer, "User1", commands.ChaincodeQuery{
		ChannelID: "testchannel",
		Name:      "mycc",
		Ctor:      `{"Args":[` + queryArgs + `]}`,
	})

	Expect(err).NotTo(HaveOccurred())
	Eventually(sess, time.Minute).Should(gexec.Exit(1))
	Expect(sess.Err).To(gbytes.Say(".+\"" + regexp.QuoteMeta(expectedResult) + "\""))
}

func RunSimpleInvoke(n *nwo.Network, orderer *nwo.Orderer, peer *nwo.Peer, args []string) {
	invokeArgs := sliceToCLIArgs(args)

	sess, err := n.PeerUserSession(peer, "User1", commands.ChaincodeInvoke{
		ChannelID: "testchannel",
		Orderer:   n.OrdererAddress(orderer, nwo.ListenPort),
		Name:      "mycc",
		Ctor:      `{"Args":[` + invokeArgs + `]}`,
		PeerAddresses: []string{
			n.PeerAddress(n.Peer("Org1", "peer0"), nwo.ListenPort),
			n.PeerAddress(n.Peer("Org2", "peer1"), nwo.ListenPort),
		},
		WaitForEvent: true,
	})
	Expect(err).NotTo(HaveOccurred())
	Eventually(sess, time.Minute).Should(gexec.Exit(0))
	Expect(sess.Err).To(gbytes.Say("Chaincode invoke successful. result: status:200"))
}

func RunSimpleBadInvoke(n *nwo.Network, orderer *nwo.Orderer, peer *nwo.Peer, args []string) {
	invokeArgs := sliceToCLIArgs(args)

	sess, err := n.PeerUserSession(peer, "User1", commands.ChaincodeInvoke{
		ChannelID: "testchannel",
		Orderer:   n.OrdererAddress(orderer, nwo.ListenPort),
		Name:      "mycc",
		Ctor:      `{"Args":[` + invokeArgs + `]}`,
		PeerAddresses: []string{
			n.PeerAddress(n.Peer("Org1", "peer0"), nwo.ListenPort),
			n.PeerAddress(n.Peer("Org2", "peer1"), nwo.ListenPort),
		},
		WaitForEvent: true,
	})
	Expect(err).NotTo(HaveOccurred())
	Eventually(sess, time.Minute).Should(gexec.Exit(1))
	Expect(sess.Err).To(gbytes.Say("Error: endorsement failure during invoke. response: status:500.*"))
}

func sliceToCLIArgs(args []string) string {
	for index, el := range args {
		args[index] = "\"" + el + "\""
	}

	return strings.Join(args, ",")
}
