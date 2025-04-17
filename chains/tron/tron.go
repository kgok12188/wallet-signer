package tron

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gin-gonic/gin"
	"github.com/mr-tron/base58"
	"log"
	"wallet-signer/models"
)

type TronSignParams struct {
	From   *string `json:"from,omitempty"`
	TxData *string `json:"raw_data_hex,omitempty"`
}

func SignTx(context *gin.Context) {
	params := TronSignParams{}
	err := context.BindJSON(&params)
	if err != nil {
		context.JSON(400, gin.H{})
		return
	}
	var rawTx []byte
	if params.TxData != nil && *params.TxData != "" {
		rawTx, err = hexutil.Decode(fmt.Sprintf("0x%s", *params.TxData))
		if err != nil {
			context.JSON(400, gin.H{})
			return
		}
	}
	dbAddress := models.Address{
		Addr: *params.From,
	}
	err = models.GetOneAddress(&dbAddress, *params.From, "tron")
	if err != nil {
		context.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}
	if dbAddress.PrivateKey == "" {
		context.JSON(400, gin.H{
			"error": "not_found_private_key",
		})
	}
	privateKey, _ := crypto.HexToECDSA(dbAddress.PrivateKey)
	hash := sha256.Sum256(rawTx)
	signature, err := crypto.Sign(hash[:], privateKey)
	if err != nil {
		context.JSON(400, gin.H{
			"error": err.Error(),
		})
	}
	context.JSON(200, gin.H{
		"data": hex.EncodeToString(signature),
	})
	return
}

func GetAddress(c *gin.Context) {
	address, privateKeyBytes := GenerateKeyPair()
	dbAddr := models.Address{
		Addr:       address,
		PrivateKey: privateKeyBytes,
		ChainId:    "tron",
	}
	err := models.AddNewAddress(&dbAddr)
	if err != nil {
		log.Println("get_tron_address", err)
	} else {
		fmt.Printf("address: %s,%s", address, privateKeyBytes)
		c.JSON(200, gin.H{
			"address": address,
		})
	}
}

func GenerateKeyPair() (b5, pk string) {
	privateKey, _ := crypto.GenerateKey()
	privateKeyBytes := crypto.FromECDSA(privateKey)
	publicKey := privateKey.Public()
	publicKeyECDSA, _ := publicKey.(*ecdsa.PublicKey)
	address := crypto.PubkeyToAddress(*publicKeyECDSA).Hex()
	address = "41" + address[2:]
	addb, _ := hex.DecodeString(address)
	firstHash := sha256.Sum256(addb)
	secondHash := sha256.Sum256(firstHash[:])
	secret := secondHash[:4]
	addb = append(addb, secret...)
	return base58.Encode(addb), hexutil.Encode(privateKeyBytes)[2:]
}
