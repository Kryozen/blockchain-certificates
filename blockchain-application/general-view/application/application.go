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
	"crypto/sha256"
	"encoding/hex"
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
	log.Println("============ Avvio applicazione certificati ============")

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
	
	seller-contract := network.GetContract("contract1")
	certificator-contract := network.GetContract("contract2")
	
	// INIZIALIZZAZIONE DEL BUFFER DI INPUT
	reader := bufio.NewReader(os.Stdin)
	
	
	// INIZIALIZZAZIONE DELLA BLOCKCHAIN
	result , err := certificator-contract.EvaluateTransaction("GetAllCertificates")
	
	// GetAllCertificates restituisce un array di asset o di json?
	var assets []Asset
	err = json.Unmarshal(result, &assets)
	if err != nil {
		log.Println("--> Inizializzo il ledger")
		_, err = certificator-contract.SubmitTransaction("InitLedger")
		if err != nil {
			log.Fatalf("Failed to Submit transaction: %v", err)
			return 
		}
	} else {
		log.Println("I dati sono già presenti.")
	}
	
	var string admin_pwd = ""
	var bool choice = false
	for {
		fmt.Println("====================================")
		fmt.Println("Selezionare la tipologia di utente:\n")
		fmt.Println("1. Certificatore")
		fmt.Println("2. Venditore")
		fmt.Println()
		op, _ := reader.ReadString('\n')
		op = strings.Replace(op, "\n", "", -1)
		
		if op == "1" {
			fmt.Println("====================================")
			fmt.Print("Inserisci la password (Lascia campo vuoto per uscire): ")
			pwd, _ := reader.ReadString('\n')
			pwd = strings.replace(pwd, "\n", "", -1)
			
			if pwd == "":
				break
			
			admin_pwd = pwd
			h := sha256.New()
			h.write([]byte(pwd))
			pwd = hex.EncodeToString(h.Sum(nil))
			
			if pwd != certificator-contract.getPwd() {
				admin_pwd = ""
				fmt.Println("Password errata.\n")
			} else {
				choice = true
				break
			}
		}
	}
		
	// SCELTA DELL'UTENTE
	if admin {
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
					
					result, err := certificator-contract.EvaluateTransaction("GetAllAssets")
					if err != nil {
						log.Fatalf("Errore nella transazione: %v\n", err)
					}
					fmt.Println(string(result))
				case "2":
					fmt.Println("====================================")
					fmt.Println("Visualizzando le richieste di certificazione in sospeso")
					
					result, err := certificator-contract.EvaluateTransaction("GetProductsPending")
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
							_, err = certificator-contract.SubmitTransaction("EvaluateProduct", id, "true")
							break
						} else if strings.ToLower(esito) == "n" {
							_, err = certificator-contract.SubmitTransaction("EvaluateProduct", id, "false")
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
					
					_, err = certificator-contract.SubmitTransaction("RenewCertificate", id)
					if err != nil {
						log.Fatalf("Errore nella transazione: %v\n", err)
					}
					fmt.Println("Certificato ", id, " rinnovato con successo.")
				case "5":
					fmt.Println("====================================")
					fmt.Print("Inserisci l'ID del certificato da annullare: ")
					
					id, _ := reader.ReadString('\n')
					id = strings.Replace(id, "\n", "", -1)
					
					_, err = certificator-contract.SubmitTransaction("InvalidateCertificate", id)
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
	} else {
		for {
			fmt.Println("====================================")
			fmt.Print("Selezionare l'operazione\n")
			fmt.Print("1. Richiedere il certificato per un prodotto\n")
			fmt.Print("2. Visualizzare i certificati esistenti\n")
			fmt.Print("3. Verificare validità di un certificato\n")
			fmt.Print("4. Richiedere il rinnovo di un certificato\n")
			fmt.Print("5. Visualizza i dettagli di un certificato\n")
			fmt.Print("6. Exit\n")
			fmt.Println("====================================")
			op, _ := reader.ReadString('\n')
			op = strings.Replace(op, "\n", "", -1)
			
			switch op {
				case "1":
					// Richiesta certificato
					fmt.Println("====================================")
					fmt.Print("Inserire il proprio nominativo: ")
					owner, _ := reader.ReadString('\n')
					owner = strings.Replace(owner, "\n", "", -1)
					
					fmt.Print("Inserire il nome del prodotto: ")
					product, _ := reader.ReadString('\n')
					product = strings.Replace(product, "\n", "", -1)
					
					fmt.Print("Inserire la tipologia di certificato richiesta: ")
					certType, _ := reader.ReadString('\n')
					certType = strings.Replace(certType, "\n", "", -1)
					
					id, err := seller-contract.SubmitTransaction("SubmitProduct", owner, product, certType)
					if err != nil {
						log.Fatalf("Errore nella transazione: %v\n", err)
						break
					}
					log.Println("Transazione SubmitProduct eseguita correttamente!\n")
					log.Println("ID della richiesta: ", string(id))
				case "2":
					// Visualizzazione di tutti i certificati
					fmt.Println("====================================")
					fmt.Println("Caricando tutti i certificati...")
					result , err := seller-contract.EvaluateTransaction("GetAllCertificates")
					if err != nil {
						log.Fatalf("Errore nella transazione: %v\n", err)
						break
					}
					fmt.Println(string(result))
				case "3":
					// Verificare validità
					fmt.Println("====================================")
					fmt.Print("Inserire l'id del certificato di cui si vuole verificare la validità: ")
					id, _ := reader.ReadString('\n')
					id = strings.Replace(id, "\n", "", -1)
					
					valid, err := seller-contract.EvaluateTransaction("VerifyCertificate", id)
					if err != nil {
						log.Fatalf("Errore nella transazione: %v\n", err)
						break
					}
					if string(valid) == "true" {
						fmt.Println("Il certificato è valido.")
					} else {
						fmt.Println("Il certificato non è valido.")
					}
					
				case "4":
					// Richiesta rinnovo
					fmt.Println("====================================")
					fmt.Print("Inserire l'id del certificato di cui si vuole richiedere il rinnovo: ")
					id, _ := reader.ReadString('\n')
					id = strings.Replace(id, "\n", "", -1)
					
					_, err := seller-contract.SubmitTransaction("RenewRequest", id)
					if err != nil {
						log.Fatalf("Errore nella transazione: %v\n", err)
						break
					}
					fmt.Println("La richiesta di rinnovo è in attesa di approvazione.")
				case "5":
					// Visualizzazione dettagli asset
					fmt.Println("====================================")
					fmt.Print("Inserire l'id del certificato di cui si vogliono visualizzare i dettagli: ")
					id, _ := reader.ReadString('\n')
					id = strings.Replace(id, "\n", "", -1)
					
					asset, err := seller-contract.EvaluateTransaction("ReadAsset", id)
					if err != nil {
						log.Fatalf("Errore nella transazione: %v\n", err)
						break
					}
					fmt.Println(string(asset))
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
