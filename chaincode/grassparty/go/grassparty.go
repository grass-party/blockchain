/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

/*
 * The sample smart contract for documentation topic:
 * Writing Your First Blockchain Application
 */

package main

/* Imports
 * 4 utility libraries for formatting, handling bytes, reading and writing JSON, and string manipulation
 * 2 specific Hyperledger Fabric specific libraries for Smart Contracts
 */
import (
	"encoding/json"
	"fmt"
//	"crypto"
	"crypto/rsa"
	"crypto/rand"
//	"crypto/md5"
	"crypto/x509"
//	"crypto/aes"
//	"crypto/cipher"
	"encoding/pem"
//	"encoding/base64"
	"strings"
	"errors"
//	"bytes"
//	"strconv"
//	"net/smtp"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	sc "github.com/hyperledger/fabric/protos/peer"
)

// Define the Smart Contract structure
type SmartContract struct {
}

// Define the car structure, with 4 properties.  Structure tags are used by encoding/json library
type Agenda struct {
	Hash   string `json:"hash"`
	A      int `json:"a"`
	B      int `json:"b"`
	C      int `json:"c"`
	D      int `json:"d"`
	E      int `json:"e"`
	Voted  map[string] bool
}

func InitAgenda(hash string) *Agenda {
	var a Agenda
	a.Hash = hash
	a.A = 0
	a.B = 0
	a.C = 0
	a.D = 0
	a.E = 0
	a.Voted = map[string]bool{}

	return &a
}

type Account struct {
	Pub   string `json:"pub"`
}

type VoteField struct {
	Account_id string `json:"account_id"`
	Agenda_id string `json:"agenda_id"`
	Vote_num string `json:"vote_num"`
	Sign string `json:"sign"`
}

func Unpad(src []byte) ([]byte, error) {
	length := len(src)
	unpadding := int(src[length-1])

	if unpadding > length {
		return nil, errors.New("unpad error. This could happen when incorrect encryption key is used")
	}

	return src[:(length - unpadding)], nil
}

/*
 * The Init method is called when the Smart Contract "grassparty" is instantiated by the blockchain network
 * Best practice is to have any Ledger initialization in separate function -- see initLedger()
 */
func (s *SmartContract) Init(APIstub shim.ChaincodeStubInterface) sc.Response {
	return shim.Success(nil)
}

/*
 * The Invoke method is called as a result of an application request to run the Smart Contract "grassparty"
 * The calling application program has also specified the particular smart contract function to be called, with arguments
 */
func (s *SmartContract) Invoke(APIstub shim.ChaincodeStubInterface) sc.Response {

	// Retrieve the requested Smart Contract function and arguments
	function, args := APIstub.GetFunctionAndParameters()
	// Route to the appropriate handler function to interact with the ledger appropriately
	if function == "initLedger" {
		return s.initLedger(APIstub)
	} else if function == "register" {
		return s.register(APIstub, args)
	} else if function == "getAccount" {
		return s.getAccount(APIstub, args)
	} else if function == "setAgenda" {
		return s.setAgenda(APIstub, args)
	} else if function == "getAgenda" {
		return s.getAgenda(APIstub, args)
	} else if function == "getSymmetricKey" {
		return s.getSymmetricKey(APIstub, args)
	} else if function == "vote" {
		return s.vote(APIstub, args)
	}

	return shim.Error("Invalid Smart Contract function name.")
}

func (s *SmartContract) initLedger(APIstub shim.ChaincodeStubInterface) sc.Response {

	return shim.Success(nil)
}

func (s *SmartContract) register(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	accountAsBytes, _ := APIstub.GetState(args[0])

	var account Account

	if accountAsBytes != nil {
		json.Unmarshal(accountAsBytes, &account)
		return shim.Error(args[0] + " is Existing User ID")
	} else {
		// public key type test
		s := strings.Replace(args[1], `\n`, "\n", -1)
		block, _ := pem.Decode([]byte(s))

		if block == nil  || block.Type != "PUBLIC KEY" {
			return shim.Error("failed to parse PEM block containing the public key :" + s + ":")
		}

		_, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return shim.Error("failed to parse DER encoded public key: " + err.Error() + " :" + s + ":")
		}
		// public key type test end

		account = Account{Pub: args[1]}

		accountAsBytes, _ := json.Marshal(account)
		APIstub.PutState(args[0], accountAsBytes)

		return shim.Success(nil)
	}
}

func (s *SmartContract) getAccount(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	accountAsBytes, _ := APIstub.GetState(args[0])

	if accountAsBytes != nil {
		return shim.Success(accountAsBytes)
	} else {
		return shim.Error("Invalid Account")
	}
}

func (s *SmartContract) setAgenda(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	agendaAsBytes, _ := APIstub.GetState(args[0])

	if agendaAsBytes != nil {
		return shim.Error(string(agendaAsBytes)+" is Existing Agenda ID")
	} else {
		agenda := InitAgenda(args[1])
		agendaAsBytes, _ := json.Marshal(agenda)
		APIstub.PutState(args[0], agendaAsBytes)

		return shim.Success(agendaAsBytes)
	}

}

func (s *SmartContract) getAgenda(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	agendaAsBytes, _ := APIstub.GetState(args[0])

	if agendaAsBytes != nil {
		return shim.Success(agendaAsBytes)
	} else {
		return shim.Error("Invalid Agenda")
	}
}

func (s *SmartContract) getSymmetricKey(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	accountAsBytes, _ := APIstub.GetState(args[0])

	if accountAsBytes != nil {
		account := Account{}
		json.Unmarshal(accountAsBytes, &account)

		s := strings.Replace(account.Pub, `\n`, "\n", -1)
		block, _ := pem.Decode([]byte(s))

		if block == nil  || block.Type != "PUBLIC KEY" {
			return shim.Error("failed to parse PEM block containing the public key")
		}

		pub, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return shim.Error("failed to parse DER encoded public key: " + err.Error())
		}
		pubkey, _ := pub.(*rsa.PublicKey)

		// !!!!!!!!!!! 나중에 바꿔야함 !!!!!!!!!!!!!!!
		key := "thisissymmetric0"

		EncryptedKey, err := rsa.EncryptPKCS1v15(rand.Reader, pubkey, []byte(key))

		if err != nil {
			return shim.Error("Encrypt Fail")
		}

		return shim.Success(EncryptedKey)
	} else {
		return shim.Error("Invalid Account")
	}
}

func (s *SmartContract) vote(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 4 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	/*
	ciphertext, err := base64.StdEncoding.DecodeString(args[0])
	if err != nil {
		return shim.Error("Base64 Decoding Error")
	}

	if len([]byte(ciphertext))%aes.BlockSize != 0 {
		return shim.Error("Crypted text must multiply by key"+args[0])
	}

	key := "thisissymmetric0"
	iv := "1234567890123456"
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return shim.Error("Error AES Symmetric Key object making")
	}

	vote_data_json := make([]byte, len(ciphertext))
	mode := cipher.NewCBCDecrypter(block, []byte(iv))

	mode.CryptBlocks(vote_data_json, ciphertext)

	vote_data_json, _ = Unpad(vote_data_json)
//	vote_data_json = []byte(strings.Replace(string(vote_data_json), `"`, "\"", -1))
//	vote_data_json = []byte("{\"account_id\":\"account_id\",\"agenda_id\":\"agenda_id\",\"vote_num\":\"C\",\"sign\":\"alkdfhklshflkjroiuweoi32h3248\"}")

	vote_data := VoteField{}
	json.Unmarshal(vote_data_json, &vote_data)

	account_id := vote_data.Account_id
	agenda_id := vote_data.Agenda_id
	vote_num := vote_data.Vote_num
//	sign := vote_data.Sign
	*/

	account_id := args[0]
	agenda_id := args[1]
	vote_num := args[2]

	//	유저 확인 및 투표했었는지 확인 필요
	accountAsBytes, _ := APIstub.GetState(account_id)
	if accountAsBytes != nil {
		// Account 사용가능하도록 바꾸기
		account := Account{}
		json.Unmarshal(accountAsBytes, &account)

		// agenda 가져오기
		agendaAsBytes, _ := APIstub.GetState(agenda_id)
		if agendaAsBytes != nil {
			agenda := Agenda{}
			json.Unmarshal(agendaAsBytes, &agenda)

/*
			// sign 검증하는 루틴 있어야함
			hashtest := md5.New()
			hashtest.Write([]byte(agenda_id+vote_num))
			digest := hashtest.Sum(nil)
			var h2 crypto.Hash
			err = rsa.VerifyPKCS1v15(pubKey, h2, digest, []byte(sign))

			if err != nil {
				return shim.Error("verify fail")
			}
*/
			// 투표했었는지 확인
			var isVoted bool = agenda.Voted[account_id]
			if isVoted != true {
				// 투표 기록
				agenda.Voted[account_id] = true

				// 투표
				switch string(vote_num) {
				case "A":
					agenda.A += 1
				case "B":
					agenda.B += 1
				case "C":
					agenda.C += 1
				case "D":
					agenda.D += 1
				case "E":
					agenda.E += 1
				default:
					fmt.Printf("unknown voting %s.", vote_num)
				}

				agendaAsBytes, _ = json.Marshal(agenda)
				APIstub.PutState(string(agenda_id), agendaAsBytes)
			} else {
				return shim.Error("This Account Already Voted")
			}
		} else {
			return shim.Error("Invalid Agenda"+args[0])
		}
	} else {
		return shim.Error("Invalid Account"+args[0]+":"+account_id)
	}

	/*
	*/


/*
	mail send
*/
	// Account Pub에서 email 받아내야 함. 아니면 확인 원할때만 email 받도록
/*
	var UserMail string = "raynear@gmail.com"
	var AcceptMail bool = true

	if AcceptMail == true {
		auth := smtp.PlainAuth("", "grasspartykr@gmail.com", "rmfotmvkxl", "smtp.gmail.com")
		from := "grasspartykr@gmail.com"
		to := []string{UserMail} // 복수 수신자 가능

		// 메시지 작성
		headerSubject := "Subject: Vote Record\r\n"
		headerBlank := "\r\n"
		body := "You vote "+args[1]+"\r\n"
		msg := []byte(headerSubject + headerBlank + body)

		// 메일 보내기
		err := smtp.SendMail("smtp.gmail.com:587", auth, from, to, msg)
		if err != nil {
			panic(err)
		}
	}
*/
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

