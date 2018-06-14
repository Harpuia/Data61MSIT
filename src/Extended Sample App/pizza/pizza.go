/******************************************************************************************************************
* File:pizza.go
* Project: MSIT-SE Studio Project (Data61)
* Copyright: Team Unchained
* Versions:
*   1.0 March 2018 - Initial implementation by Dongliang Zhou
*	2.0 March 2018 - Added access control logic by Dongliang Zhou
*
* Description: This is the smart contract created manually to implement the pizza BPMN. 
*
* External Dependencies: Hyperledger fabric library
*
******************************************************************************************************************/


package main
import (
	"bytes"
	"strings"
	"encoding/json"
	"encoding/pem"
	"crypto/x509"
	"fmt"
	"strconv"
	//"github.com/hyperledger/fabric/core/chaincode/lib/cid"
    "github.com/hyperledger/fabric/core/chaincode/shim"
    "github.com/hyperledger/fabric/protos/peer"
)


// Define the Smart Contract structure
type SmartContract struct {
}

// Define the order structure, with 5 properties.  Structure tags are used by encoding/json library
type Order struct {
	OrderId string `json:"id"`
	Item  string `json:"item"`
	Customer string `json:"customer"`
	Deliverer  string `json:"deliverer"`
	Status string `json:"status"`
}

/*
 * The Init method is called when the Smart Contract "pizza" is instantiated by the blockchain network
 * Best practice is to have any Ledger initialization in separate function -- see initLedger()
 */
func (s *SmartContract) Init(APIstub shim.ChaincodeStubInterface) peer.Response {
	return shim.Success(nil)
}

/*
 * The Invoke method is called as a result of an application request to run the Smart Contract "pizza"
 * The calling application program has also specified the particular smart contract function to be called, with arguments
 */
func (s *SmartContract) Invoke(APIstub shim.ChaincodeStubInterface) peer.Response {

	// Retrieve the requested Smart Contract function and arguments
	function, args := APIstub.GetFunctionAndParameters()

	// Creator info
    fmt.Println("Invoke Start Here")

    // GetCreator returns marshaled serialized identity of the client
    creatorByte,_:= APIstub.GetCreator()
    certStart := bytes.IndexAny(creatorByte, "-----BEGIN")
    if certStart == -1 {
       return shim.Error("No certificate found")
    }
    certText := creatorByte[certStart:]
    bl, _ := pem.Decode(certText)
    if bl == nil {
       return shim.Error("Could not decode the PEM structure")
    }

    cert, err := x509.ParseCertificate(bl.Bytes)
    if err != nil {
       return shim.Error("ParseCertificate failed")
    }
    uname:=cert.Subject.CommonName
    domainStart := strings.Index(uname,"@")
    if domainStart == -1 {
    	return shim.Error("Could not parce certificate domain")
    }
    domain := uname[domainStart+1:]

    // Do whatever you need with certificate

	// Route to the appropriate handler function to interact with the ledger appropriately
	if function == "queryOrder" {
		return s.queryOrder(APIstub, args)
	} else if function == "initLedger" {
		if domain == "restaurant.example.com" {
			return s.initLedger(APIstub)
		} else {
			return shim.Error(domain+" is not allowed to call "+function)
		}
	} else if function == "createOrder" {
		if domain == "customer.example.com" {
			return s.createOrder(APIstub, args)
		} else {
			return shim.Error(domain+" is not allowed to call "+function)
		}
	} else if function == "queryAllOrders" {		
		return s.queryAllOrders(APIstub)
	} else if function == "confirmOrder" {		
		if domain == "restaurant.example.com" {
			return s.confirmOrder(APIstub, args)
		} else {
			return shim.Error(domain+" is not allowed to call "+function)
		}
	} else if function == "cancelOrder" {		
		if domain == "restaurant.example.com" {
			return s.cancelOrder(APIstub, args)
		} else {
			return shim.Error(domain+" is not allowed to call "+function)
		}
	} else if function == "assignDeliverer" {
		if domain == "restaurant.example.com" {
			return s.assignDeliverer(APIstub, args)
		} else {
			return shim.Error(domain+" is not allowed to call "+function)
		}
	} else if function == "deliverOrder" {		
		if domain == "deliverer.example.com" {
			return s.deliverOrder(APIstub, args)
		} else {
			return shim.Error(domain+" is not allowed to call "+function)
		}
	}

	return shim.Error("Invalid Smart Contract function name.")
}

func (s *SmartContract) queryOrder(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	orderAsBytes, _ := APIstub.GetState(args[0])
	return shim.Success(orderAsBytes)
}

func (s *SmartContract) initLedger(APIstub shim.ChaincodeStubInterface) peer.Response {
	orders := []Order{
		Order{OrderId: "ORDER0", Item: "Cheese Pizza", Customer: "Leo", Deliverer: "John", Status: "Delivered"},
	}

	i := 0
	for i < len(orders) {
		fmt.Println("i is ", i)
		orderAsBytes, _ := json.Marshal(orders[i])
		APIstub.PutState("ORDER"+strconv.Itoa(i), orderAsBytes)
		fmt.Println("Added", orders[i])
		i = i + 1
	}

	return shim.Success(nil)
}

func (s *SmartContract) createOrder(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {
	fmt.Println("creating! ");
	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}

	var order = Order{OrderId: args[0], Item: args[1], Customer: args[2], Deliverer: "N/A", Status: "Ordered"}

	orderAsBytes, _ := json.Marshal(order)
	APIstub.PutState(args[0], orderAsBytes)

	return shim.Success(nil)
}

func (s *SmartContract) queryAllOrders(APIstub shim.ChaincodeStubInterface) peer.Response {

	startKey := "ORDER0"
	endKey := "ORDER999"

	resultsIterator, err := APIstub.GetStateByRange(startKey, endKey)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing QueryResults
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"Key\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResponse.Key)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Record\":")
		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	fmt.Printf("- queryAllOrders:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}

func (s *SmartContract) confirmOrder(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	orderAsBytes, _ := APIstub.GetState(args[0])
	order := Order{}

	json.Unmarshal(orderAsBytes, &order)
	if order.Status != "Ordered" {
		return shim.Error("This order cannot be confirmed")
	}

	order.Status = "Confirmed"

	orderAsBytes, _ = json.Marshal(order)
	APIstub.PutState(args[0], orderAsBytes)

	return shim.Success(nil)
}

func (s *SmartContract) cancelOrder(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	orderAsBytes, _ := APIstub.GetState(args[0])
	order := Order{}

	json.Unmarshal(orderAsBytes, &order)
	if order.Status == "Delivered" {
		return shim.Error("This order cannot be cancelled.")
	}
	
	order.Status = "Cancelled"

	orderAsBytes, _ = json.Marshal(order)
	APIstub.PutState(args[0], orderAsBytes)

	return shim.Success(nil)
}
func (s *SmartContract) assignDeliverer(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	orderAsBytes, _ := APIstub.GetState(args[0])
	order := Order{}

	json.Unmarshal(orderAsBytes, &order)
	if order.Status != "Confirmed" {
		return shim.Error("Cannot assign deliverer to this order")
	}
	
	order.Deliverer = args[1]
	order.Status = "OutForDelivery"

	orderAsBytes, _ = json.Marshal(order)
	APIstub.PutState(args[0], orderAsBytes)

	return shim.Success(nil)
}
func (s *SmartContract) deliverOrder(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	orderAsBytes, _ := APIstub.GetState(args[0])
	order := Order{}

	json.Unmarshal(orderAsBytes, &order)
	if order.Status != "OutForDelivery" {
		return shim.Error("This order cannot be delivered")
	}
	
	order.Status = "Delivered"

	orderAsBytes, _ = json.Marshal(order)
	APIstub.PutState(args[0], orderAsBytes)

	return shim.Success(nil)
}
// The main function is only relevant in unit test mode. Only included here for completeness.
func main() {

	// Create a new Smart Contract
	err := shim.Start(new(SmartContract))
	if err != nil {
		fmt.Printf("Error creating new Smart Contract: %s", err)
	}
}
