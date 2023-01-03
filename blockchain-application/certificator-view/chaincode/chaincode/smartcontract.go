package chaincode

import (
	"encoding/json"
	"fmt"
	"strings"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"time"
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
//ID is obtained concatenating owner + product + certType

// InitLedger adds a base set of assets to the ledger
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	assets := []Asset{
		{ID: "mattiapandorodop", Owner: "Mattia", Product: "Pandoro", CertType: "D.O.P.", ExpireDate: "25/12/2023", Renew: false},
		{ID: "simonecotechinoigp", Owner: "Simone", Product: "Cotechino", CertType: "I.G.P.", ExpireDate: "01/01/2024", Renew: false},
		{ID: "antonellaaglianicodoc", Owner: "Antonella", Product: "Aglianico beneventano", CertType: "D.O.C.", ExpireDate: "09/04/2023", Renew: false},
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

// CreateAsset issues a new asset to the world state with given details.
func (s *SmartContract) CreateAsset(ctx contractapi.TransactionContextInterface, owner string, product string, certType string, expireDate string) error {
	//Calculating ID
	id := strings.ToLower(owner + product + strings.Replace(certType, ".", "", -1))
	id = strings.Replace(id, " ", "", -1)
	
	exists, err := s.AssetExists(ctx, id)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the asset %s already exists", id)
	}

	asset := Asset{
		ID:             id,
		Owner:		owner,
		Product:	product,
		CertType:	certType,
		ExpireDate:	expireDate,
		Renew:		false,
	}
	
	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, assetJSON)
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

// UpdateAsset updates an existing asset in the world state with provided parameters.
func (s *SmartContract) UpdateAsset(ctx contractapi.TransactionContextInterface, id string, owner string, product string, certType string, expireDate string) error {
	exists, err := s.AssetExists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the asset %s does not exist", id)
	}
	
	//Calculating ID
	new_id := strings.ToLower(owner + product + strings.Replace(certType, ".", "", -1))
	new_id = strings.Replace(new_id, " ", "", -1)

	// overwriting original asset with new asset
	asset := Asset{
		ID:             new_id,
		Owner:		owner,
		Product:	product,
		CertType:	certType,
		ExpireDate:	expireDate,
	}
	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	s.DeleteAsset(ctx, id)
	return ctx.GetStub().PutState(new_id, assetJSON)
}

// DeleteAsset deletes an given asset from the world state.
func (s *SmartContract) DeleteAsset(ctx contractapi.TransactionContextInterface, id string) error {
	exists, err := s.AssetExists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the asset %s does not exist", id)
	}

	return ctx.GetStub().DelState(id)
}

// AssetExists returns true when asset with given ID exists in world state
func (s *SmartContract) AssetExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	assetJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return assetJSON != nil, nil
}

// TransferAsset updates the owner field of asset with given id in world state.
func (s *SmartContract) TransferAsset(ctx contractapi.TransactionContextInterface, id string, newOwner string) error {
	asset, err := s.ReadAsset(ctx, id)
	if err != nil {
		return err
	}

	asset.Owner = newOwner
	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, assetJSON)
}

// GetAllAssets returns all assets found in world state
func (s *SmartContract) GetAllAssets(ctx contractapi.TransactionContextInterface) ([]*Asset, error) {
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
		assets = append(assets, &asset)
	}

	return assets, nil
}

// GetProductsPending returns all the products waiting for certification
func (s *SmartContract) GetProductsPending(ctx contractapi.TransactionContextInterface) ([]*Asset, error) {
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
		if asset.ExpireDate == "1980-01-01" || asset.Renew {
			assets = append(assets, &asset)
		}
	}

	return assets, nil
}

// EvaluateProduct approves or refuses a product in queue for a certificate
func (s *SmartContract) EvaluateProduct(ctx contractapi.TransactionContextInterface, id string, evaluation bool) error {
	var asset *Asset
	asset, err := s.ReadAsset(ctx, id)
	if err != nil {
		return err
	}
	
	if asset.ExpireDate != "1980-01-01" {
		return fmt.Errorf("Tale ID corrisponde ad un certificato già valido.")
	}
	
	if evaluation == true {
		currentTime := time.Now().Add(time.Hour * 24 * 365)
		asset.ExpireDate = string(currentTime.Format("dd-MM-yyyy"))
		
		assetJSON, err := json.Marshal(asset)
		if err != nil {
			return err
		}
	
		return ctx.GetStub().PutState(id, assetJSON)
	} else {
		return s.DeleteAsset(ctx, id)
	}	
}

// RenewCertificates modifies the expiration date of the certificate
func (s *SmartContract) RenewCertificate(ctx contractapi.TransactionContextInterface, id string) error {
	var asset *Asset
	asset, err := s.ReadAsset(ctx, id)
	if err != nil {
		return err
	}
	
	if asset.Renew != true {
		return fmt.Errorf("the asset %s does not have a pending request.",id)
	}
	
	oldExpireDate, err := time.Parse("dd-MM-yyyy", asset.ExpireDate)
	if err != nil {
		return err
	}
	
	today := time.Now()
	
	var newest time.Time
	
	if today.After(oldExpireDate) {
		newest = today
	} else {
		newest = oldExpireDate
	}
	
	currentTime := newest.AddDate(1,0,0)
	asset.ExpireDate = string(currentTime.Format("dd-MM-yyyy"))
	asset.Renew = false
	
	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, assetJSON)
}

// InvalidateCertificate changes the expirationDate of a certificate to the day before the operation is submitted
func (s *SmartContract) InvalidateCertificate(ctx contractapi.TransactionContextInterface, id string) error {
	var asset *Asset
	asset, err := s.ReadAsset(ctx, id)
	if err != nil {
		return err
	}
	
	yesterday := time.Now().Add(-24 * time.Hour)
	asset.ExpireDate = string(yesterday.Format("dd-MM-yyyy"))
	
	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, assetJSON)
}
