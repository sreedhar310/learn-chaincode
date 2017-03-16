/*
Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements.  See the NOTICE file
distributed with this work for additional information
regarding copyright ownership.  The ASF licenses this file
to you under the Apache License, Version 2.0 (the
"License"); you may not use this file except in compliance
with the License.  You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing,
software distributed under the License is distributed on an
"AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
KIND, either express or implied.  See the License for the
specific language governing permissions and limitations
under the License.
*/

package main

import (
	"errors"
	"fmt"
	"strconv"
	"encoding/json"
	"strings"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

var accountIndexStr = "_accountindex"			

type Account struct{
	AccountNo string `json:"accountno"`	
	LegalEntity string `json:"legalentity"`
	Currency string `json:"currency"`				
	Balance string `json:"balance"`
}

// ============================================================================================================================
// Main
// ============================================================================================================================
func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

// ============================================================================================================================
// Init - reset all the things
// ============================================================================================================================
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	var Aval int
	var err error

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1")
	}

	// Initialize the chaincode
	Aval, err = strconv.Atoi(args[0])
	if err != nil {
		return nil, errors.New("Expecting integer value for asset holding")
	}

	// Write the state to the ledger
	err = stub.PutState("test_key", []byte(strconv.Itoa(Aval)))				//making a test var "abc", I find it handy to read/write to it right away to test the network
	if err != nil {
		return nil, err
	}
	
	var empty []string
	jsonAsBytes, _ := json.Marshal(empty)								//marshal an emtpy array of strings to clear the index
	err = stub.PutState(accountIndexStr, jsonAsBytes)
	if err != nil {
		return nil, err
	}
	
	return nil, nil
}

// ============================================================================================================================
// Invoke - Our entry point for Invocations
// ============================================================================================================================
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("invoke is running " + function)

	// Handle different functions
	if function == "init" {													//initialize the chaincode state, used as reset
		return t.Init(stub, "init", args)
	} else if function == "delete" {										//deletes an entity from its state
		return t.Delete(stub, args)												//lets make sure all open trades are still valid
	} else if function == "write" {											//writes a value to the chaincode state
		return t.Write(stub, args)
	} else if function == "init_account" {									//create a new account
		return t.init_account(stub, args)
	} else if function == "transfer_balance" {									
		return t.transfer_balance(stub, args)										
	}
	fmt.Println("invoke did not find func: " + function)					//error

	return nil, errors.New("Received unknown function invocation")
}

// ============================================================================================================================
// Query - Our entry point for Queries
// ============================================================================================================================
func (t *SimpleChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("query is running " + function)

	// Handle different functions
	if function == "read" {													//read a variable
		return t.read(stub, args)
	}
	fmt.Println("query did not find func: " + function)						//error

	return nil, errors.New("Received unknown function query")
}

// ============================================================================================================================
// Read - read a variable from chaincode state
// ============================================================================================================================
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

// ============================================================================================================================
// Delete - remove a key/value pair from state
// ============================================================================================================================
func (t *SimpleChaincode) Delete(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1")
	}
	
	name := args[0]
	err := stub.DelState(name)													//remove the key from chaincode state
	if err != nil {
		return nil, errors.New("Failed to delete state")
	}

	//get the marble index
	accountsAsBytes, err := stub.GetState(accountIndexStr)
	if err != nil {
		return nil, errors.New("Failed to get marble index")
	}
	var accountIndex []string
	json.Unmarshal(accountsAsBytes, &accountIndex)								//un stringify it aka JSON.parse()
	
	//remove marble from index
	for i,val := range accountIndex{
		fmt.Println(strconv.Itoa(i) + " - looking at " + val + " for " + name)
		if val == name{															//find the correct account
			fmt.Println("found account")
			accountIndex = append(accountIndex[:i], accountIndex[i+1:]...)			//remove it
			for x:= range accountIndex{											//debug prints...
				fmt.Println(string(x) + " - " + accountIndex[x])
			}
			break
		}
	}
	jsonAsBytes, _ := json.Marshal(accountIndex)									//save new index
	err = stub.PutState(accountIndexStr, jsonAsBytes)
	return nil, nil
}

// ============================================================================================================================
// Write - write variable into chaincode state
// ============================================================================================================================
func (t *SimpleChaincode) Write(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var name, value string // Entities
	var err error
	fmt.Println("running write()")

	if len(args) != 2 {
		return nil, errors.New("Incorrect number of arguments. Expecting 2. name of the variable and value to set")
	}

	name = args[0]														
	value = args[1]
	err = stub.PutState(name, []byte(value))					
	if err != nil {
		return nil, err
	}
	return nil, nil
}

// ============================================================================================================================
// Init account - create a new account, store into chaincode state
// ============================================================================================================================
func (t *SimpleChaincode) init_account(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var err error

	//       0        1      2      3
	// "accountNo", "bob", "USD", "3500"
	if len(args) != 4 {
		return nil, errors.New("Incorrect number of arguments. Expecting 4")
	}

	//input sanitation
	fmt.Println("- start init acount")
	if len(args[0]) <= 0 {
		return nil, errors.New("1st argument must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		return nil, errors.New("2nd argument must be a non-empty string")
	}
	if len(args[2]) <= 0 {
		return nil, errors.New("3rd argument must be a non-empty string")
	}
	if len(args[3]) <= 0 {
		return nil, errors.New("3rd argument must be a non-empty string")
	}

	accountNo := args[0]

	legalEntity := strings.ToLower(args[1])

	currency := args[2]

	ammount, err := strconv.ParseFloat(args[3],64)
	if err != nil {
		return nil, errors.New("4rd argument must be a numeric string")
	}

	//check if account already exists
	accountAsBytes, err := stub.GetState(accountNo)
	if err != nil {
		return nil, errors.New("Failed to get account number")
	}
	res := Account{}
	json.Unmarshal(accountAsBytes, &res)
	if res.AccountNo == accountNo{
		fmt.Println("This account arleady exists: " + accountNo)
		fmt.Println(res);
		return nil, errors.New("This account arleady exists")			
	}
	amountStr := strconv.FormatFloat(ammount, 'E', -1, 64)
	//build the account json string manually
	str := `{"accountno": "` + accountNo + `", "legalentity": "` + legalEntity + `", "currency": "` + currency + `", "balance": "` + amountStr + `"}`
	err = stub.PutState(accountNo, []byte(str))							
	if err != nil {
		return nil, err
	}
		
	//get the account index
	accountsAsBytes, err := stub.GetState(accountIndexStr)
	if err != nil {
		return nil, errors.New("Failed to get account index")
	}
	var accountIndex []string
	json.Unmarshal(accountsAsBytes, &accountIndex)							
	
	//append
	accountIndex = append(accountIndex, accountNo)						
	fmt.Println("! account index: ", accountIndex)
	jsonAsBytes, _ := json.Marshal(accountIndex)
	err = stub.PutState(accountIndexStr, jsonAsBytes)						

	fmt.Println("- end init account")
	return nil, nil
}

// ============================================================================================================================
// transfer the balance between accounts
// ============================================================================================================================
func (t *SimpleChaincode) transfer_balance(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var err error
	var newAmountA, newAmountB float64
	//       0           1         2
	// "accountA", "accountB", "100.20"
	if len(args) < 3 {
		return nil, errors.New("Incorrect number of arguments. Expecting 3")
	}
	
	fmt.Println("- start transfer_balance")
	fmt.Println(args[0] + " to " + args[1])

	amount,err := strconv.ParseFloat(args[2], 64)
	if err != nil {
		return nil, errors.New("3rd argument must be a numeric string")
	}

	accountAAsBytes, err := stub.GetState(args[0])
	if err != nil {
		return nil, errors.New("Failed to get the first account")
	}
	resA := Account{}
	json.Unmarshal(accountAAsBytes, &resA)										//un stringify it aka JSON.parse()
	
	accountBAsBytes, err := stub.GetState(args[1])
	if err != nil {
		return nil, errors.New("Failed to get the second account")
	}
	resB := Account{}
	json.Unmarshal(accountBAsBytes, &resB)											
	
	BalanceA,err := strconv.ParseFloat(resA.Balance, 64)
	if err != nil {
		return nil, err
	}
	BalanceB,err := strconv.ParseFloat(resB.Balance, 64)
	if err != nil {
		return nil, err
	}

	if (BalanceA - amount) < 0 {
		return nil, errors.New(args[0] + " doesn't have enough balance to complete transaction")
	}

	newAmountA = BalanceA - amount
	newAmountB =  BalanceB + amount
	newAmountStrA := strconv.FormatFloat(newAmountA, 'E', -1, 64)
	newAmountStrB := strconv.FormatFloat(newAmountB, 'E', -1, 64)

	resA.Balance = newAmountStrA
	resB.Balance = newAmountStrB

	jsonAAsBytes, _ := json.Marshal(resA)
	err = stub.PutState(args[0], jsonAAsBytes)								
	if err != nil {
		return nil, err
	}

	jsonBAsBytes, _ := json.Marshal(resB)
	err = stub.PutState(args[1], jsonBAsBytes)								
	if err != nil {
		return nil, err
	}
	
	fmt.Println("- end transfer_balance")
	return nil, nil
}