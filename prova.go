/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"fmt"
	"crypto/sha256"
	//io/ioutil"
	"log"
	//"os"
	//"path/filepath"
	//"bufio"
	//"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	//"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
	"encoding/hex"
	//"strings"
)



func main() {
	log.Println("============ Avvio applicazione certificati per venditori ============")
	s:="prova"
	h:= sha256.New()
	h.Write([]byte(s))
	bs:= h.Sum(nil)
	id := hex.EncodeToString(bs)
	fmt.Printf(id)
	
}
