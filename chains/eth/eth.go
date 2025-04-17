package eth

import (
	"crypto/ecdsa"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gin-gonic/gin"
	"log"
	"math/big"
	"strings"
	"wallet-signer/models"
)

func GetAddress(c *gin.Context) {
	privateKey, _ := crypto.GenerateKey()
	privateKeyBytes := crypto.FromECDSA(privateKey)
	publicKey := privateKey.Public()
	publicKeyECDSA, _ := publicKey.(*ecdsa.PublicKey)
	address := crypto.PubkeyToAddress(*publicKeyECDSA).Hex()
	address = strings.ToLower(address)
	dbAddr := models.Address{
		Addr:       address,
		PrivateKey: fmt.Sprintf("%x", privateKeyBytes),
		ChainId:    "ETH",
	}
	err := models.AddNewAddress(&dbAddr)
	if err != nil {
		log.Println("get_eth_address", err)
	} else {
		fmt.Printf("address: %s,%x", address, privateKeyBytes)
		c.JSON(200, gin.H{
			"address": address,
		})
	}
}

type EthSignParams struct {
	Nonce    uint64   `json:"nonce"`
	ChainID  *big.Int `json:"chainID"`
	From     string   `json:"from"`
	To       string   `json:"to"`
	Amount   *big.Int `json:"amount"`
	GasLimit uint64   `json:"gasLimit"`
	GasPrice *big.Int `json:"gasPrice"`
	Data     string   `json:"data"`
}

func SignTx(context *gin.Context) {
	params := EthSignParams{}
	err := context.BindJSON(&params)
	if err != nil {
		context.JSON(400, gin.H{})
		return
	}
	var data []byte
	if params.Data != "" {
		data, err = hexutil.Decode(params.Data)
		if err != nil {
			context.JSON(400, gin.H{})
			return
		}
	}

	sign, err := Sign(params.Nonce, params.ChainID, params.From, params.To, params.Amount, params.GasLimit, params.GasPrice, data)
	if err != nil {
		context.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}
	context.JSON(200, gin.H{
		"data": sign,
	})
}

func Sign(nonce uint64, chainID *big.Int, from string, to string, amount *big.Int, gasLimit uint64, gasPrice *big.Int, data []byte) (string, error) {
	dbAddress := models.Address{
		Addr: from,
	}
	err := models.GetOneAddress(&dbAddress, from, "ETH")
	if err != nil {
		log.Println("get_eth_address", err)
		return "", err
	}
	if dbAddress.PrivateKey == "" {
		return "", errors.New("not found private key")
	}

	tx := types.NewTransaction(nonce, common.HexToAddress(to), amount, gasLimit, gasPrice, data)

	privateKey, err := crypto.HexToECDSA(dbAddress.PrivateKey)
	if err != nil {
		return "", err
	}
	tx, err = types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		return "", err
	}
	rawTxBytes, err := tx.MarshalBinary()
	if err != nil {
		return "", err
	}
	encode := hexutil.Encode(rawTxBytes)
	return encode, nil
}
