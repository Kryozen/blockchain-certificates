package chaincode

import (
	"encoding/json"
	"fmt"
	"crypto/sha256"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
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
}
//ID is calculated through the function SHA256 given the string owner+product+certtype+expiredate

// InitLedger adds a base set of assets to the ledger
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	assets := []Asset{
		{ID: "1fe656a7513296b13285a3d9a2a963e24e6461aa3603142fdf502b4b6cfcf90e", Owner: "Mattia", Product: "Pandoro", CertType: "D.O.P.", ExpireDate: "25/12/2023"},
		{ID: "88ce84afa390ec97d5debc1c988f598cc27e4d87c5fdd38f09876c976ffb2885", Owner: "Simone", Product: "Cotechino", CertType: "I.G.P.", ExpireDate: "01/01/2024"},
		{ID: "0ef8efcc68d6b365e4678fa4d6cbddaae93879242cf5a76522704ca45692dae4", Owner: "Antonella", Product: "Aglianico beneventano", CertType: "D.O.C.", ExpireDate: "09/04/2023"},
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

// SubmitProduct adds a product to the list of the products waiting for a certification
func (s *SmartContract) SubmitProduct(ctx contractapi.TransactionContextInterface, owner string, product string, certType string) error {
	//Calculating SHA256 for the certificate
	h := sha256.New()
	h.Write([]byte(owner+product+certType))
	
	id := h.Sum(nil)
	expireDate := nil
	
	exists, err := s.AssetExists(id)
	if err != nil {
		return err, fmt.Errorf("failed to read from world state: %v", err)
	}
	if exists {
		return nil, fmt.Errorf("this product is already submitted for certification.")
	} else {
		asset := Asset{
			ID:		id,
			Owner:		owner,
			Product:	product,
			CertType:	certType,
			ExpireDate:	expireDate,
		}
		
		assetJSON, err := json.Marshal(asset)
		if err != nil {
			return err
		}

		return ctx.GetStub().PutState(id, assetJSON)
	}
}

// VerifyCertificate returns true if the certificate given exists and is still valid
func (s *SmartContract) VerifyCertificate(ctx contractapi.TransactionContextInterface, hashCode string) error {
	assetJson, err := s.ReadAsset(hashCode)
	if err != nil {
		return nil, err
	}
	if asset == nil {
		return false, nil
	}
	
	var asset Asset
	err = json.Unmarshal(assetJson, &asset)
	if err != nil {
		return nil, err
	}
	
	if time.Parse("DD-MM-YYYY", asset.ExpireDate) >= time.Parse("DD-MM-YYYY", time.Now()) {
		return true, nil
	}
	return false, nil
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
