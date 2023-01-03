/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"bufio"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
	"encoding/json"
	"strings"
)

type Asset struct {
	ID     	        string `json:"ID"`
	Owner  	        string `json:"Owner"`
	Product		string `json:"Product"`
	CertType	string `json:"CertType"`
	ExpireDate	string `json:"ExpireDate"`
	Renew		bool   `json:"Renew"`
}

func main() {
	log.Println("============ Avvio applicazione certificati per certificatori ============")

	err := os.Setenv("DISCOVERY_AS_LOCALHOST", "true")
	if err != nil {
		log.Fatalf("Error setting DISCOVERY_AS_LOCALHOST environemnt variable: %v", err)
	}
	
	wallet, err := gateway.NewFileSystemWallet("wallet")
	if err != nil {
		log.Fatalf("Failed to create wallet: %v", err)
	}

	if !wallet.Exists("appUser") {
		err = populateWallet(wallet)
		if err != nil {
			log.Fatalf("Failed to populate wallet contents: %v", err)
		}
	}

	ccpPath := filepath.Join(
		"..",
		"..",
		"..",
		"network-setup",
		"organizations",
		"peerOrganizations",
		"org1.example.com",
		"connection-org1.yaml",
	)

	gw, err := gateway.Connect(
		gateway.WithConfig(config.FromFile(filepath.Clean(ccpPath))),
		gateway.WithIdentity(wallet, "appUser"),
	)
	if err != nil {
		log.Fatalf("Failed to connect to gateway: %v", err)
	}
	defer gw.Close()

	network, err := gw.GetNetwork("mychannel")
	if err != nil {
		log.Fatalf("Failed to get network: %v", err)
	}

	contract := network.GetContract("contract2")
	
	// INIZIALIZZAZIONE DEL BUFFER DI INPUT
	reader := bufio.NewReader(os.Stdin)
	
	
	// INIZIALIZZAZIONE DELLA BLOCKCHAIN
	result , err := contract.EvaluateTransaction("GetAllCertificates")
	
	// GetAllCertificates restituisce un array di asset o di json?
	var assets []Asset
	err = json.Unmarshal(result, &assets)
	if err != nil {
		log.Println("--> Inizializzo il ledger")
		_, err = contract.SubmitTransaction("InitLedger")
		if err != nil {
			log.Fatalf("Failed to Submit transaction: %v", err)
			return 
		}
	} else {
		log.Println("I dati sono già presenti.")
	}
		
	// SCELTA DELL'UTENTE
	for {
		fmt.Println("====================================")
		fmt.Print("Selezionare l'operazione\n")
		fmt.Print("1. Visualizza tutti i certificati e le richieste di certificazione\n")
		fmt.Print("2. Visualizza le richieste di certificazione in sospeso\n")
		fmt.Print("3. Approva una richiesta di certificazione\n")
		fmt.Print("4. Rinnova un certificato\n")
		fmt.Print("5. Annulla un certificato\n")
		fmt.Print("6. Exit\n")
		fmt.Println("====================================")
		op, _ := reader.ReadString('\n')
		op = strings.Replace(op, "\n", "", -1)
		
		switch op {
			case "1":
				fmt.Println("====================================")
				fmt.Println("Visualizzando tutti gli elementi")
				
				result, err := contract.EvaluateTransaction("GetAllAssets")
				if err != nil {
					log.Fatalf("Errore nella transazione: %v\n", err)
				}
				fmt.Println(string(result))
			case "2":
				fmt.Println("====================================")
				fmt.Println("Visualizzando le richieste di certificazione in sospeso")
				
				result, err := contract.EvaluateTransaction("GetProductsPending")
				if err != nil {
					log.Fatalf("Errore nella transazione: %v\n", err)
				}
				fmt.Println(string(result))
			case "3":
				fmt.Println("====================================")
				fmt.Print("Inserisci l'ID della richiesta da valutare: ")
				id, _ := reader.ReadString('\n')
				id = strings.Replace(id, "\n", "", -1)
				
				fmt.Print("Inserisci l'esito della valutazione (Y/N): ")
				esito, _ := reader.ReadString('\n')
				esito = strings.Replace(esito, "\n", "", -1)
				
				var err error
				for {
					if strings.ToLower(esito) == "y" {
						_, err = contract.SubmitTransaction("EvaluateProduct", id, "true")
						break
					} else if strings.ToLower(esito) == "n" {
						_, err = contract.SubmitTransaction("EvaluateProduct", id, "false")
						break
					}
					fmt.Println("Inserire una valutazione valida (Y/N).")
				}
				if err != nil {
					log.Fatalf("Errore nella transazione: %v\n.", err)
				}
				fmt.Println("Richiesta ", id," valutata con successo.")
			case "4":
				fmt.Println("====================================")
				fmt.Print("Inserisci l'ID del certificato da rinnovare: ")
				
				id, _ := reader.ReadString('\n')
				id = strings.Replace(id, "\n", "", -1)
				
				_, err = contract.SubmitTransaction("RenewCertificate", id)
				if err != nil {
					log.Fatalf("Errore nella transazione: %v\n", err)
				}
				fmt.Println("Certificato ", id, " rinnovato con successo.")
			case "5":
				fmt.Println("====================================")
				fmt.Print("Inserisci l'ID del certificato da annullare: ")
				
				id, _ := reader.ReadString('\n')
				id = strings.Replace(id, "\n", "", -1)
				
				_, err = contract.SubmitTransaction("InvalidateCertificate", id)
				if err != nil {
					log.Fatalf("Errore nella transazione: %v\n", err)
				}
				fmt.Println("Certificato ", id, " annullato con successo.")
			default:
				if op != "6" {
					fmt.Println("====================================")
					fmt.Println("Inserire un'opzione valida (1-6).")
					break
				}
			}
			if op == "6" {
				break
		}
	}

	log.Println("============ Terminazione dell'applicazione ============")
}

func populateWallet(wallet *gateway.Wallet) error {
	log.Println("============ Populating wallet ============")
	credPath := filepath.Join(
		"..",
		"..",
		"..",
		"network-setup",
		"organizations",
		"peerOrganizations",
		"org1.example.com",
		"users",
		"User1@org1.example.com",
		"msp",
	)

	certPath := filepath.Join(credPath, "signcerts", "User1@org1.example.com-cert.pem")
	// read the certificate pem
	cert, err := ioutil.ReadFile(filepath.Clean(certPath))
	if err != nil {
		return err
	}

	keyDir := filepath.Join(credPath, "keystore")
	// there's a single file in this dir containing the private key
	files, err := ioutil.ReadDir(keyDir)
	if err != nil {
		return err
	}
	if len(files) != 1 {
		return fmt.Errorf("keystore folder should have contain one file")
	}
	keyPath := filepath.Join(keyDir, files[0].Name())
	key, err := ioutil.ReadFile(filepath.Clean(keyPath))
	if err != nil {
		return err
	}

	identity := gateway.NewX509Identity("Org1MSP", string(cert), string(key))

	return wallet.Put("appUser", identity)
}
