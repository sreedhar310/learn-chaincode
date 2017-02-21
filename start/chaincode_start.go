
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

// Init resets all the things
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1")
	}

	err := stub.PutState("test_key", []byte(args[0]))
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// Invoke isur entry point to invoke a chaincode function
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("invoke is running " + function)

	// Handle different functions
	if function == "init" {
		return t.Init(stub, "init", args)
	} else if function == "write" {
		return t.write(stub, args)
	}
	fmt.Println("invoke did not find func: " + function)

	return nil, errors.New("Received unknown function invocation: " + function)
}

// Query is our entry point for queries
func (t *SimpleChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("query is running " + function)

	// Handle different functions
	if function == "read" { //read a variable
		return t.read(stub, args)
	}
	fmt.Println("query did not find func: " + function)

	return nil, errors.New("Received unknown function query: " + function)
}

// write - invoke function to write key/value pair
func (t *SimpleChaincode) write(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var key, value string
	var err error
	fmt.Println("running write()")

	if len(args) != 2 {
		return nil, errors.New("Incorrect number of arguments. Expecting 2. name of the key and value to set")
	}

	key = args[0] //rename for funsies
	value = args[1]
	err = stub.PutState(key, []byte(value)) //write the variable into the chaincode state
	if err != nil {
		return nil, err
	}
	return nil, nil
}

// read - query function to read key/value pair
func (t *SimpleChaincode) read(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var key, jsonResp string
	var err error

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting name of the key to query")
	}

	key = args[0]
	valAsbytes, err := stub.GetState(key)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + key + "\"}"
		return nil, errors.New(jsonResp)
	}

	return valAsbytes, nil
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
	amount,err := strconv.ParseFloat(args[1], 64)
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
	if len(args) != 3 {
		return nil, errors.New("Incorrect number of arguments. Expecting 3")
	}
	
	fmt.Println("- start transfer money")
	fmt.Println("- from " + args[0] + " to " + args[1])

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

	err = stub.PutState(args[1], []byte(newAmountStrB))		

	if err != nil {
		return nil, err
	}
	
	fmt.Println("- transfer completed")
	return nil, nil
}