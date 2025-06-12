package main

import (
	"fmt"
	"os"
	"encoding/json"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	wutils "github.com/hyperledger-cacti/cacti/weaver/core/network/fabric-interop-cc/libs/utils/v2"
)

// SmartContract provides functions for managing arbitrary key-value pairs
type SmartContract struct {
	contractapi.Contract
}

// Defines the CRM structure
type CRM struct {
	Id         string   `json:"id"`
	Specialty  string   `json:"specialty"`
	State      bool     `json:"state"`
}

// Record Interoperation Chaincode ID on ledger
func (s *SmartContract) Init(ctx contractapi.TransactionContextInterface, ccId string) error {
	ccBytes := []byte(ccId)
	fmt.Printf("Init called. CC ID: %s\n", ccId)

	return ctx.GetStub().PutState(wutils.GetInteropChaincodeIDKey(), ccBytes)
}

// Create adds a new entry with the specified id and state
func (s *SmartContract) Create(ctx contractapi.TransactionContextInterface, id string, specialty string, state bool) error {
	crm := CRM{
		Id: id,
		Specialty: specialty,
		State: state,
	}
	crmJSON, err := json.Marshal(crm)
	if err != nil {
		return err
	}
	fmt.Printf("Create called. Id: %s Specialty %s State: %t\n", id, specialty, state)

	return ctx.GetStub().PutState(id, crmJSON)
}

// Read returns the value of the entry with the specified id
func (s *SmartContract) Read(ctx contractapi.TransactionContextInterface, id string) (*CRM, error) {
	relayAccessCheck, err := wutils.CheckAccessIfRelayClient(ctx.GetStub())
	if err != nil {
		return nil, err
	}
	if !relayAccessCheck {
		return nil, fmt.Errorf("Illegal access by relay")
	}
	fmt.Printf("Relay access check passed.")

	crmJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("Failed to read id '%s' from world state. %s", id, err.Error())
	}
	if crmJSON == nil {
		return nil, fmt.Errorf("The CRM %s does not exist.", id)
	}

	var crm CRM
	err = json.Unmarshal(crmJSON, &crm)
	if err != nil {
		return nil, err
	}

	return &crm, nil
}

// Check validates a CRM entry based on the specified id and returns the doctors Specialty
func (s *SmartContract) Check(ctx contractapi.TransactionContextInterface, id string) (string, error) {
	relayAccessCheck, err := wutils.CheckAccessIfRelayClient(ctx.GetStub())
	if err != nil {
		return "", err
	}
	if !relayAccessCheck {
		return "", fmt.Errorf("Illegal access by relay")
	}
	fmt.Printf("Relay access check passed.")

	crmJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return "", fmt.Errorf("Failed to read id '%s' from world state. %s", id, err.Error())
	}
	if crmJSON == nil {
		return "", fmt.Errorf("The CRM %s does not exist.", id)
	}

	var crm CRM
	err = json.Unmarshal(crmJSON, &crm)
	if err != nil {
		return "", err
	}

	res := ""
	if crm.State {
		res = crm.Specialty
	} else {
		res = "Invalid"
	}

	return res, nil
}

func main() {
	chaincode, err := contractapi.NewChaincode(new(SmartContract))

	if err != nil {
		fmt.Printf("Error create CRM chaincode: %s", err.Error())
		return
	}

	_, ok := os.LookupEnv("EXTERNAL_SERVICE")
	if ok {
		server := &shim.ChaincodeServer{
				CCID:    os.Getenv("CHAINCODE_CCID"),
				Address: os.Getenv("CHAINCODE_ADDRESS"),
				CC:      chaincode,
				TLSProps: shim.TLSProperties{
										Disabled: true,
									},
		}
		// Start the chaincode external server
		err = server.Start()
	} else {
		err = chaincode.Start()
	}
	if err != nil {
		fmt.Printf("Error starting CRM chaincode: %s", err.Error())
	}
}
