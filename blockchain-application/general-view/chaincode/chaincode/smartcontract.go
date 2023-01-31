package chaincode

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"time"
	"crypto/sha256"
	"encoding/hex"
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

var pwd_hash string

// InitLedger adds a base set of assets to the ledger
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	pwd_hash = "ef92b778bafe771e89245b89ecbc08a44a4e166c06659911881f383d4473e94f"
	assets := []Asset{
		{ID: "mattiapandorodop", Owner: "Mattia", Product: "Pandoro", CertType: "D.O.P.", ExpireDate: "2023-12-25", Renew: false},
		{ID: "simonecotechinoigp", Owner: "Simone", Product: "Cotechino", CertType: "I.G.P.", ExpireDate: "2024-01-01", Renew: false},
		{ID: "antonellaaglianicodoc", Owner: "Antonella", Product: "Aglianico beneventano", CertType: "D.O.C.", ExpireDate: "2023-04-09", Renew: false},
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
func (s *SmartContract) CreateAsset(ctx contractapi.TransactionContextInterface, pwd string, owner string, product string, certType string, expireDate string) error {
	
	//Checking pwd
	if s.CheckPwd(pwd) == false {
		return fmt.Errorf("the password is not correct")
	}
		
	//Calculating ID
	str1 := owner + product + certType
	h := sha256.New()
	h.Write([]byte(str1))
	bs := h.Sum(nil)
	id := hex.EncodeToString(bs)
	
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
func (s *SmartContract) UpdateAsset(ctx contractapi.TransactionContextInterface, pwd string, id string, owner string, product string, certType string, expireDate string) error {
	
	//Checking pwd
	if s.CheckPwd(pwd) == false {
		return fmt.Errorf("the password is not correct")
	}

	exists, err := s.AssetExists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the asset %s does not exist", id)
	}
	
	//Calculating ID
	str1 := owner + product + certType
	h := sha256.New()
	h.Write([]byte(str1))
	bs := h.Sum(nil)
	new_id := hex.EncodeToString(bs)
	fmt.Println(id)

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

	s.DeleteAsset(ctx, pwd, id)
	return ctx.GetStub().PutState(new_id, assetJSON)
}

// DeleteAsset deletes an given asset from the world state.
func (s *SmartContract) DeleteAsset(ctx contractapi.TransactionContextInterface, pwd string, id string) error {

	//Checking pwd
	if s.CheckPwd(pwd) == false {
		return fmt.Errorf("the password is not correct")
	}
	
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
func (s *SmartContract) TransferAsset(ctx contractapi.TransactionContextInterface, pwd string, id string, newOwner string) error {
	//Checking pwd
	if s.CheckPwd(pwd) == false {
		return fmt.Errorf("the password is not correct")
	}
	
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
func (s *SmartContract) GetAllAssets(ctx contractapi.TransactionContextInterface, pwd string) ([]*Asset, error) {
	//Checking pwd
	if s.CheckPwd(pwd) == false {
		return nil, fmt.Errorf("the password is not correct")
	}
	
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
func (s *SmartContract) GetProductsPending(ctx contractapi.TransactionContextInterface, pwd string) ([]*Asset, error) {
	//Checking pwd
	if s.CheckPwd(pwd) == false {
		return nil, fmt.Errorf("the password is not correct")
	}
	
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
func (s *SmartContract) EvaluateProduct(ctx contractapi.TransactionContextInterface, pwd string, id string, evaluation bool) error {
	//Checking pwd
	if s.CheckPwd(pwd) == false {
		return fmt.Errorf("the password is not correct")
	}
	
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
				
		asset.ExpireDate = currentTime.Format("2006-01-02")
		
		assetJSON, err := json.Marshal(asset)
		if err != nil {
			return err
		}
	
		return ctx.GetStub().PutState(id, assetJSON)
	} else {
		return s.DeleteAsset(ctx, pwd, id)
	}	
}

// RenewCertificates modifies the expiration date of the certificate
func (s *SmartContract) RenewCertificate(ctx contractapi.TransactionContextInterface, pwd string, id string) error {
	//Checking pwd
	if s.CheckPwd(pwd) == false {
		return fmt.Errorf("the password is not correct")
	}
	
	var asset *Asset
	asset, err := s.ReadAsset(ctx, id)
	if err != nil {
		return err
	}
	
	if asset.Renew != true {
		return fmt.Errorf("the asset %s does not have a pending request.",id)
	}
	
	
	if err != nil {
		return err
	}
	
	oldExpireDate, err := time.Parse("2006-01-02", asset.ExpireDate)
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
	asset.ExpireDate = currentTime.Format("2006-01-02")
	asset.Renew = false
	
	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, assetJSON)
}

// InvalidateCertificate changes the expirationDate of a certificate to the day before the operation is submitted
func (s *SmartContract) InvalidateCertificate(ctx contractapi.TransactionContextInterface, pwd string, id string) error {
	//Checking pwd
	if s.CheckPwd(pwd) == false {
		return fmt.Errorf("the password is not correct")
	}
	
	var asset *Asset
	asset, err := s.ReadAsset(ctx, id)
	if err != nil {
		return err
	}
	
	yesterday := time.Now().Add(-24 * time.Hour)
	asset.ExpireDate = yesterday.Format("2006-01-02")
	
	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, assetJSON)
}

func (s *SmartContract) CheckPwd(pwd string) bool {
	h := sha256.New()
	h.Write([]byte(pwd))
	hash := hex.EncodeToString(h.Sum(nil))
	
	return hash == pwd_hash
}

// SubmitProduct adds a product to the list of the products waiting for a certification
func (s *SmartContract) SubmitProduct(ctx contractapi.TransactionContextInterface, owner string, product string, certType string) (string, error) {
	//Calculating SHA256 for the certificate
	
	str1 := owner + product + certType
	h := sha256.New()
	h.Write([]byte(str1))
	bs := h.Sum(nil)
	id := hex.EncodeToString(bs)

	
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
