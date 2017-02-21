
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

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
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
	err = stub.PutState("testKey", []byte(strconv.Itoa(Aval)))				//making a test var "testKey", I find it handy to read/write to it right away to test the network
	if err != nil {
		return nil, err
	}
	
	return nil, nil
}

// ============================================================================================================================
// Run - Our entry point for Invocations - [LEGACY] obc-peer 4/25/2016
// ============================================================================================================================
func (t *SimpleChaincode) Run(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("run is running " + function)
	return t.Invoke(stub, function, args)
}

// ============================================================================================================================
// Invoke - Our entry point for Invocations
// ============================================================================================================================
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("invoke is running " + function)

	// Handle different functions
	if function == "init" {													
		return t.Init(stub, "init", args)
	} else if function == "delete" {										
		res, err := t.Delete(stub, args)
		return res, err
	} else if function == "write" {								
		return t.Write(stub, args)
	} else if function == "init_amount" {									
		return t.init_amount(stub, args)
	} else if function == "transfer" {										
		return t.transfer(stub, args)
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
		return t.Read(stub, args)
	}
	fmt.Println("query did not find func: " + function)						//error

	return nil, errors.New("Received unknown function query")
}

// ============================================================================================================================
// Read - read a variable from chaincode state
// ============================================================================================================================
func (t *SimpleChaincode) Read(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
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

	name = args[0]															//rename for funsies
	value = args[1]
	err = stub.PutState(name, []byte(value))								//write the variable into the chaincode state
	if err != nil {
		return nil, err
	}
	return nil, nil
}

// ============================================================================================================================
// Init Amount - set the initial amount for user
// ============================================================================================================================
func (t *SimpleChaincode) init_amount(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var err error

	//   0      1 
	// "bob", "200.45"
	if len(args) != 2 {
		return nil, errors.New("Incorrect number of arguments. Expecting 2")
	}

	//input sanitation
	fmt.Println("- start init amount")
	if len(args[0]) <= 0 {
		return nil, errors.New("1st argument must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		return nil, errors.New("2nd argument must be a non-empty string")
	}
	
	name := args[0]
	amount,err := strconv.ParseFloat(args[2], 64)
	if err != nil {
		return nil, errors.New("2nd argument must be a numeric string")
	}
	amountStr := strconv.FormatFloat(amount, 'E', -1, 64)
	err = stub.PutState(name, []byte(amountStr))									//store marble with id as key
	if err != nil {
		return nil, err
	}

	fmt.Println("- end init amount")
	return nil, nil
}

// ============================================================================================================================
// Transfer money from user A to user B
// ============================================================================================================================
func (t *SimpleChaincode) transfer(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var err error
	var userA, userB, jsonResp string
	var newAmountA, newAmountB float64
	
	//    0       1      2
	// "alice", "bob", "12.56"
	if len(args) < 3 {
		return nil, errors.New("Incorrect number of arguments. Expecting 3")
	}
	
	fmt.Println("- start transfer money")
	fmt.Println("from " + args[0] + " to " + args[1])

	userA = args[0]
	userB = args[1]
	amount,err := strconv.ParseFloat(args[2], 64)
	if err != nil {
		return nil, errors.New("3rd argument must be a numeric string")
	}
	amountByteA, err := stub.GetState(userA)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + userA + "\"}"
		return nil, errors.New(jsonResp)
	}	
	amountByteB, err := stub.GetState(userB)	
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + userB + "\"}"
		return nil, errors.New(jsonResp)
	}
	amountStrA := string(amountByteA[:])
	amountStrB := string(amountByteB[:])
	amountA,err := strconv.ParseFloat(amountStrA, 64)
	if err != nil {
		return nil, err
	}
	amountB,err := strconv.ParseFloat(amountStrB, 64)
	if err != nil {
		return nil, err
	}

	if (amountA - amount) < 0 {
		return nil, errors.New(args[0] + " doesn't have enough balance to complete transaction")
	} 
	newAmountA = amountA - amount
	newAmountB =  amountB + amount
	newAmountStrA := strconv.FormatFloat(newAmountA, 'E', -1, 64)
	newAmountStrB := strconv.FormatFloat(newAmountB, 'E', -1, 64)


	err = stub.PutState(args[0], []byte(newAmountStrA))		

	if err != nil {
		return nil, err
	}

	err = stub.PutState(args[0], []byte(newAmountStrB))		

	if err != nil {
		return nil, err
	}
	
	fmt.Println("- transfer completed")
	return nil, nil
}
