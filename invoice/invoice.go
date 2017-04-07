package main

import (
	"errors"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"encoding/json"
)

//==============================================================================================================================
//	 Participant types - Each participant type is mapped to an integer which we use to compare to the value stored in a
//						 user's eCert
//==============================================================================================================================

const   SUPPLIER   =  "supplier"
const   PAYER   =  "payer"
const   BUYER =  "buyer"


//==============================================================================================================================
//	 Structure Definitions
//==============================================================================================================================
//	Chaincode - A blank struct for use with Shim (A HyperLedger included go file used for get/put state
//				and other HyperLedger functions)
//==============================================================================================================================
type  SimpleChaincode struct {
}

//==============================================================================================================================
//	Vehicle - Defines the structure for a car object. JSON on right tells it what JSON fields to map to
//			  that element when reading a JSON object into the struct e.g. JSON make -> Struct Make.
//==============================================================================================================================
type Invoice struct {
	InvoiceId        string `json:"invoiceid"`
	Amount           string `json:"amount"`
	Currency         string `json:"currency"`
	Supplier         string    `json:"supplier"`
	Payer            string `json:"payer"`
	DueDate          string   `json:"duedate"`
	Status           int    `json:"status"`
	Buyer            string `json:"buyer"`
	Discount         string `json:"discount"`

}


//==============================================================================================================================
//	V5C Holder - Defines the structure that holds all the v5cIDs for vehicles that have been created.
//				Used as an index when querying all vehicles.
//==============================================================================================================================

type Invoice_Holder struct {
	Invoices 	[]string `json:"invoices"`
}


//==============================================================================================================================
//	Init Function - Called when the user deploys the chaincode
//==============================================================================================================================
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {

	//Args
	//				0              1             2            3            4            5
	//			test_user0      supplier    test_user1      payer      test_user2     buyer

	var invoiceIDs Invoice_Holder

	bytes, err := json.Marshal(invoiceIDs)

    if err != nil { return nil, errors.New("Error creating Invoice_Holder record") }

	err = stub.PutState("invoiceIDs", bytes)
	if err != nil { return nil, errors.New("Error putting state with invoiceIDs") }

	for i:=0; i < len(args); i=i+2 {
		t.add_particants(stub, args[i], args[i+1])
	}

	return nil, nil
}

//==============================================================================================================================
//	 General Functions
//==============================================================================================================================

func (t *SimpleChaincode) get_role(stub shim.ChaincodeStubInterface, name string) ([]byte, error) {

	role, err := stub.GetState(name)

	if err != nil { return nil, errors.New("Couldn't retrieve role for user " + name) }

	return role, nil
}


func (t *SimpleChaincode) add_particants(stub shim.ChaincodeStubInterface, name string, role string) ([]byte, error) {


	err := stub.PutState(name, []byte(role))

	if err != nil {
		return nil, errors.New("Error storing user " + name + " role: " + role)
	}

	return nil, nil

}

//==============================================================================================================================
//	 get_caller - Retrieves the username of the user who invoked the chaincode.
//				  Returns the username as a string.
//==============================================================================================================================

func (t *SimpleChaincode) get_username(stub shim.ChaincodeStubInterface) (string, error) {

    username, err := stub.ReadCertAttribute("username");
	if err != nil { return "", errors.New("Couldn't get attribute 'username'. Error: " + err.Error()) }
	return string(username), nil
}


//==============================================================================================================================
//	 get_caller_data - Calls the get_ecert and check_role functions and returns the ecert and role for the
//					 name passed.
//==============================================================================================================================

func (t *SimpleChaincode) get_caller_data(stub shim.ChaincodeStubInterface) (string, string, error){

	user, err := t.get_username(stub)

	role, err := t.get_role(stub,user);

    if err != nil { return "", "", err }

	return user, string(role), nil
}

//==============================================================================================================================
//	 retrieve_invoice
//==============================================================================================================================
func (t *SimpleChaincode) retrieve_invoice(stub shim.ChaincodeStubInterface, invoiceId string) (Invoice, error) {

	var inv Invoice

	bytes, err := stub.GetState(invoiceId);

	if err != nil {	fmt.Printf("RETRIEVE_INVOICE: Failed to invoke invoice id: %s", err); return inv, errors.New("RETRIEVE_INVOICE: Error retrieving invoice with invoice Id = " + invoiceId) }

	err = json.Unmarshal(bytes, &inv);

    if err != nil {	fmt.Printf("RETRIEVE_INVOICE: Corrupt invoice record "+string(bytes)+": %s", err); return inv, errors.New("RETRIEVE_INVOICE: Corrupt invoice record"+string(bytes))	}

	return inv, nil
}

//==============================================================================================================================
// save_changes - Writes to the ledger the Vehicle struct passed in a JSON format. Uses the shim file's
//				  method 'PutState'.
//==============================================================================================================================
func (t *SimpleChaincode) save_changes(stub shim.ChaincodeStubInterface, inv Invoice) (bool, error) {

	bytes, err := json.Marshal(inv)

	if err != nil { fmt.Printf("SAVE_CHANGES: Error converting invoice record: %s", err); return false, errors.New("Error converting invoice record") }

	err = stub.PutState(inv.InvoiceId, bytes)

	if err != nil { fmt.Printf("SAVE_CHANGES: Error storing invoice record: %s", err); return false, errors.New("Error storing invoice record") }

	return true, nil
}

//==============================================================================================================================
//	 Router Functions
//==============================================================================================================================
//	Invoke - Called on chaincode invoke. Takes a function name passed and calls that function. Converts some
//		  initial arguments passed to other things for use in the called function e.g. name -> ecert
//==============================================================================================================================
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {

	caller, role, err := t.get_caller_data(stub)

	if err != nil { return nil, errors.New("Error retrieving caller information")}


	if function == "create_invoice" {
        return t.create_invoice(stub, caller, role, args)
	} else {
        return t.ping(stub)
    } 

}
//=================================================================================================================================
//	Query - Called on chaincode query. Takes a function name passed and calls that function. Passes the
//  		initial arguments passed are passed on to the called function.
//=================================================================================================================================
func (t *SimpleChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {

	caller, role, err := t.get_caller_data(stub)
	if err != nil { fmt.Printf("QUERY: Error retrieving caller details", err); return nil, errors.New("QUERY: Error retrieving caller details: "+err.Error()) }

	if function == "get_invoice_details" {
		if len(args) != 1 { fmt.Printf("Incorrect number of arguments passed"); return nil, errors.New("QUERY: Incorrect number of arguments passed") }
		inv, err := t.retrieve_invoice(stub, args[0])
		if err != nil { fmt.Printf("QUERY: Error retrieving nvoice: %s", err); return nil, errors.New("QUERY: Error retrieving invoice "+err.Error()) }
		return t.get_invoice_details(stub, inv, caller, role)
	} else if function == "check_unique_invoice" {
		return t.check_unique_invoice(stub, args[0], caller, role)
	} else if function == "get_invoices" {
		return t.get_invoices(stub, caller, role)
	}  else if function == "read" {													//read a variable
		return t.read(stub, args)
	} else if function == "get_username" {													//read a variable
		return stub.ReadCertAttribute(args[0]);
	} else {
		return t.ping(stub)
	} 

	return nil, errors.New("Received unknown function invocation " + function)

}

//=================================================================================================================================
//	 Ping Function
//=================================================================================================================================
//	 Pings the peer to keep the connection alive
//=================================================================================================================================
func (t *SimpleChaincode) ping(stub shim.ChaincodeStubInterface) ([]byte, error) {
	return []byte("Hello, world!"), nil
}

func (t *SimpleChaincode) read(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var name, jsonResp string
	var err error

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting name of the var to query")
	}

	name = args[0]
	valAsbytes, err := stub.GetState(name)									//get the var from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + name + "\"}"
		return nil, errors.New(jsonResp)
	}

	return valAsbytes, nil													//send it onward
}

//=================================================================================================================================
//	 Create Function
//=================================================================================================================================
//	 Create Vehicle - Creates the initial JSON for the vehcile and then saves it to the ledger.
//=================================================================================================================================
func (t *SimpleChaincode) create_invoice(stub shim.ChaincodeStubInterface, caller string, role string, args []string) ([]byte, error) {
	var inv Invoice

	invId          := "\"invoiceid\":\""+args[0]+"\", "							// Variables to define the JSON
	amount         := "\"amount\":\""+args[1]+"\", "	
	currency       := "\"currency\":\"USD\", "
	supplier       := "\"supplier\":\""+caller+"\", "
	payer          := "\"payer\":\"UNDEFINED\", "
	status         := "\"status\":\"0\", "
	buyer          := "\"buyer\":\"UNDEFINED\", "
	discount       := "\"discount\":\"UNDEFINED\", "

	var invoiceId = args[0]

	invoice_json := "{"+invId+amount+currency+supplier+payer+status+buyer+discount+"}" 	// Concatenates the variables to create the total JSON object


	err := json.Unmarshal([]byte(invoice_json), &inv)							// Convert the JSON defined above into a vehicle object for go

	if err != nil { return nil, errors.New("Invalid JSON object") }

	record, err := stub.GetState(inv.InvoiceId) 								// If not an error then a record exists so cant create a new car with this V5cID as it must be unique

	if record != nil { return nil, errors.New("Invoice already exists") }

	if 	role != SUPPLIER {						

		return nil, errors.New(fmt.Sprintf("Permission Denied. create_invoice. %v === %v", role, SUPPLIER))

	}

	_, err  = t.save_changes(stub, inv)

	if err != nil { fmt.Printf("CREATE_INVOICE: Error saving changes: %s", err); return nil, errors.New("Error saving changes") }

	bytes, err := stub.GetState("invoiceIDs")

	if err != nil { return nil, errors.New("Unable to get invoiceIDs") }

	var invoiceIDs Invoice_Holder

	err = json.Unmarshal(bytes, &invoiceIDs)

	if err != nil {	return nil, errors.New("Corrupt Invoice_Holder record") }

	invoiceIDs.Invoices = append(invoiceIDs.Invoices, invoiceId)

	bytes, err = json.Marshal(invoiceIDs)

	if err != nil { fmt.Print("Error creating Invoice_Holder record") }

	err = stub.PutState("invoiceIDs", bytes)

	if err != nil { return nil, errors.New("Unable to put the state") }

	return nil, nil

}

//=================================================================================================================================
//	 Read Functions
//=================================================================================================================================
//	 get_vehicle_details
//=================================================================================================================================
func (t *SimpleChaincode) get_invoice_details(stub shim.ChaincodeStubInterface, inv Invoice, caller string, caller_affiliation string) ([]byte, error) {

	bytes, err := json.Marshal(inv)

	if err != nil { return nil, errors.New("GET_VEHICLE_DETAILS: Invalid vehicle object") }

	if 		inv.Supplier  == caller		||
			inv.Buyer	== caller	{
				return bytes, nil
	} else {
			return nil, errors.New("Permission Denied. get_invoice_details")
	}

}

//=================================================================================================================================
//	 get_vehicles
//=================================================================================================================================

func (t *SimpleChaincode) get_invoices(stub shim.ChaincodeStubInterface, caller string, role string) ([]byte, error) {
	bytes, err := stub.GetState("invoiceIDs")

	if err != nil { return nil, errors.New("Unable to get invoiceIDs") }

	var invoiceIDs Invoice_Holder

	err = json.Unmarshal(bytes, &invoiceIDs)

	if err != nil {	return nil, errors.New("Corrupt Invoice_Holder") }

	result := "["

	var temp []byte
	var inv Invoice

	for _, invoiceId := range invoiceIDs.Invoices {

		inv, err = t.retrieve_invoice(stub, invoiceId)

		if err != nil {return nil, errors.New("Failed to retrieve Invoice")}

		temp, err = t.get_invoice_details(stub, inv, caller, role)

		if err == nil {
			result += string(temp) + ","
		}
	}

	if len(result) == 1 {
		result = "[]"
	} else {
		result = result[:len(result)-1] + "]"
	}

	return []byte(result), nil
}

//=================================================================================================================================
//	 check_unique_v5c
//=================================================================================================================================
func (t *SimpleChaincode) check_unique_invoice(stub shim.ChaincodeStubInterface, invoiceId string, caller string, caller_affiliation string) ([]byte, error) {
	_, err := t.retrieve_invoice(stub, invoiceId)
	if err == nil {
		return []byte("false"), errors.New("invoice is not unique")
	} else {
		return []byte("true"), nil
	}
}

//=================================================================================================================================
//	 Main - main - Starts up the chaincode
//=================================================================================================================================
func main() {

	err := shim.Start(new(SimpleChaincode))
	if err != nil { fmt.Printf("Error starting Chaincode: %s", err) }
}