package chaincode
//Contract for sellers

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"time"
	"strings"
)

// SmartContract provides functions for managing an Asset
type SmartContract struct {
	contractapi.Contract
}

// Asset describes basic details of what makes up a simple asset
//Insert struct field in alphabetic order => to achieve determinism accross languages
// golang keeps the order when marshal to json but doesn't order automatically
type Asset struct {
	ID     	        string `json:"ID"`
	Owner  	        string `json:"Owner"`
	Product		string `json:"Product"`
	CertType	string `json:"CertType"`
	ExpireDate	string `json:"ExpireDate"`
	Renew		bool   `json:"Renew"`
}
//ID is obtained concatenating owner+product+certtype

// InitLedger adds a base set of assets to the ledger
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	assets := []Asset{
		{ID: "mattiapandorodop", Owner: "Mattia", Product: "Pandoro", CertType: "D.O.P.", ExpireDate: "2023-12-25", Renew: false},
		{ID: "simonecotechinoigp", Owner: "Simone", Product: "Cotechino", CertType: "I.G.P.", ExpireDate: "2024-01-01", Renew: false},
		{ID: "antonellaaglianicodoc", Owner: "Antonella", Product: "Aglianico beneventano", CertType: "D.O.C.", ExpireDate: "2023-09-04", Renew: false},
	}

	for _, asset := range assets {
		assetJSON, err := json.Marshal(asset)
		if err != nil {
			return err
		}

		err = ctx.GetStub().PutState(asset.ID, assetJSON)
		if err != nil {
			return fmt.Errorf("failed to put to world state. %v", err)
		}
	}

	return nil
}

// ReadAsset returns the asset stored in the world state with given id.
func (s *SmartContract) ReadAsset(ctx contractapi.TransactionContextInterface, id string) (*Asset, error) {
	assetJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if assetJSON == nil {
		return nil, fmt.Errorf("the asset %s does not exist", id)
	}

	var asset Asset
	err = json.Unmarshal(assetJSON, &asset)
	if err != nil {
		return nil, err
	}

	return &asset, nil
}

// AssetExists returns true when asset with given ID exists in world state
func (s *SmartContract) AssetExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	assetJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return assetJSON != nil, nil
}

// SubmitProduct adds a product to the list of the products waiting for a certification
func (s *SmartContract) SubmitProduct(ctx contractapi.TransactionContextInterface, owner string, product string, certType string) (string, error) {
	id := strings.ToLower(owner + product + strings.Replace(certType, ".", "", -1))
	id = strings.Replace(id, " ", "", -1)
	
	expireDate := "1980-01-01"
	
	exists, err := s.AssetExists(ctx, id)
	if err != nil {
		return "", fmt.Errorf("Errore nella ricerca di un asset esistente, %v", err)
	}
	if exists {
		return "", fmt.Errorf("La richiesta è già stata inserita")
	} else {
		asset := Asset{
			ID:		id,
			Owner:		owner,
			Product:	product,
			CertType:	certType,
			ExpireDate:	expireDate,
			Renew:		false,
		}
		
		assetJSON, err := json.Marshal(asset)
		if err != nil {
			return "", fmt.Errorf("Errore nel marshaling: %v", err)
		}

		return id, ctx.GetStub().PutState(id, assetJSON)
	}
}

// VerifyCertificate returns true if the certificate given exists and is still valid
func (s *SmartContract) VerifyCertificate(ctx contractapi.TransactionContextInterface, toVerify string) (bool, error) {
	var asset *Asset
	
	asset, err := s.ReadAsset(ctx, toVerify)
	if err != nil {
		return false, err
	}
	if asset == nil {
		return false, nil
	}
	
	assetExpireTime, err := time.Parse("2006-01-02", asset.ExpireDate)
	if err != nil {
		return false, err
	}
	currentTime := time.Now()
	
	if assetExpireTime.After(currentTime) {
		return true, nil
	}
	return false, nil
}

// GetAllCertificates returns all certificates that have already been approved (also not valid ones anymore)
func (s *SmartContract) GetAllCertificates(ctx contractapi.TransactionContextInterface) ([]*Asset, error) {
	// range query with empty string for startKey and endKey does an
	// open-ended query of all assets in the chaincode namespace.
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var assets []*Asset
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var asset Asset
		err = json.Unmarshal(queryResponse.Value, &asset)
		if err != nil {
			return nil, err
		}
		if asset.ExpireDate != "1980-01-01" {
			assets = append(assets, &asset)
		}
	}

	return assets, nil
}

// RenewRequest sets the value Renew of the asset to be true (which means pending for renew)
func (s *SmartContract) RenewRequest(ctx contractapi.TransactionContextInterface, id string) error {
	var asset *Asset
	
	asset, err := s.ReadAsset(ctx, id)
	if err != nil {
		return err
	}
	if asset == nil {
		return nil
	}
	
	if asset.ExpireDate == "1980-01-01" {
		return fmt.Errorf("Tale ID corrisponde ad un prodotto non ancora certificato.\n")
	}
	
	asset.Renew = true
	
	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, assetJSON)
}
