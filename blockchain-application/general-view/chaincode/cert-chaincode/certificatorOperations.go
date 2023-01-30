/*
SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"log"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"certificator-view/chaincode"
)

func main() {
	certificatorChaincode, err := contractapi.NewChaincode(&chaincode.SmartContract{})
	if err != nil {
		log.Panicf("Error creating certificator chaincode: %v", err)
	}

	if err := certificatorChaincode.Start(); err != nil {
		log.Panicf("Error starting certificator chaincode: %v", err)
	}
}
