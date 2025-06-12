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

// Defines the Prescription structure
type Prescription struct {
	Id         string  `json:"id"`     
	CRM        string  `json:"crm"`
	Specialty  string  `json:"specialty"`
	Hospital   string  `json:"hospital"`
	Medicine   string  `json:"medicine"`
	Validated  string  `json:"validated"`
}

// Record Interoperation Chaincode ID on ledger
func (s *SmartContract) Init(ctx contractapi.TransactionContextInterface, ccId string) error {
	ccBytes := []byte(ccId)
	fmt.Printf("Init called. CC ID: %s\n", ccId)

	return ctx.GetStub().PutState(wutils.GetInteropChaincodeIDKey(), ccBytes)
}

// Create adds a new entry with the specified id, crm, hospital and medicine
func (s *SmartContract) Create(ctx contractapi.TransactionContextInterface, id string, crm string, specialty string, hospital string, medicine string) error {
	prescription := Prescription{
		Id: id,
		CRM: crm,
		Specialty: specialty,
		Hospital: hospital,
		Medicine: medicine,
		Validated: "Not validated",
	}
	prescriptionJSON, err := json.Marshal(prescription)
	if err != nil {
		return err
	}
	fmt.Printf("Create called. Id: %s CRM: %s Specialty: %s Hospital: %s Medicine: %s\n", id, crm, specialty, hospital, medicine)

	return ctx.GetStub().PutState(id, prescriptionJSON)
}

// Read returns the value of the entry with the specified id
func (s *SmartContract) Read(ctx contractapi.TransactionContextInterface, id string) (*Prescription, error) {
	relayAccessCheck, err := wutils.CheckAccessIfRelayClient(ctx.GetStub())
	if err != nil {
		return nil, err
	}
	if !relayAccessCheck {
		return nil, fmt.Errorf("Illegal access by relay")
	}
	fmt.Printf("Relay access check passed.")

	prescriptionJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("Failed to read id '%s' from world state. %s", id, err.Error())
	}
	if prescriptionJSON == nil {
		return nil, fmt.Errorf("The Prescription %s does not exist.", id)
	}

	var prescription Prescription
	err = json.Unmarshal(prescriptionJSON, &prescription)
	if err != nil {
		return nil, err
	}

	return &prescription, nil
}

// Check validates the prescription based on the information from the remote network
func (s *SmartContract) Check(ctx contractapi.TransactionContextInterface, id string, validation string) error {
	relayAccessCheck, err := wutils.CheckAccessIfRelayClient(ctx.GetStub())
	if err != nil {
		return err
	}
	if !relayAccessCheck {
		return fmt.Errorf("Illegal access by relay")
	}
	fmt.Printf("Relay access check passed.")

	prescriptionJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return fmt.Errorf("Failed to read id '%s' from world state. %s", id, err.Error())
	}
	if prescriptionJSON == nil {
		return fmt.Errorf("The Prescription %s does not exist.", id)
	}

	var prescription Prescription
	err = json.Unmarshal(prescriptionJSON, &prescription)
	if err != nil {
		return err
	}

	if prescription.Specialty != validation {
		prescription.Validated = "Invalid"
	} else {
		prescription.Validated = "Valid"
	}

	prescriptionJSON, err = json.Marshal(prescription)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, prescriptionJSON)
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
