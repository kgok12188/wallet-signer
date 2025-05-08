package sol

import (
	"encoding/base64"
	"encoding/json"
	"github.com/blocto/solana-go-sdk/common"
	"github.com/blocto/solana-go-sdk/program/associated_token_account"
	"github.com/blocto/solana-go-sdk/program/system"
	"github.com/blocto/solana-go-sdk/program/token"
	"github.com/blocto/solana-go-sdk/types"
	"github.com/gin-gonic/gin"
	"github.com/mr-tron/base58"
	"log"
	"strconv"
	"wallet-signer/models"
)

func GetAddress(c *gin.Context) {
	account := types.NewAccount()
	dbAddr := models.Address{
		Addr:       account.PublicKey.String(),
		PrivateKey: base58.Encode(account.PrivateKey),
		ChainId:    "SOL",
	}
	err := models.AddNewAddress(&dbAddr)
	if err != nil {
		c.JSON(400, gin.H{})
		log.Println("get_sol_address", err)
	} else {
		c.JSON(200, gin.H{
			"address": dbAddr.Addr,
		})
	}
}

func SignTx(context *gin.Context) {
	params := SignParams{}
	err := context.BindJSON(&params)
	if err != nil {
		log.Println("SignTx", err)
		context.JSON(400, gin.H{})
		return
	}
	if params.FeePayer == "" || params.Instructions == nil || len(params.Instructions) == 0 || params.Blockhash == "" {
		context.JSON(400, gin.H{})
		return
	}
	var Instructions []types.Instruction
	var fromAccounts = make(map[string]*types.Account)
	for _, instruction := range params.Instructions {
		amount, err := strconv.ParseUint(instruction.Amount.String(), 10, 64)
		if err != nil {
			log.Println("金额错误", err)
			context.JSON(400, gin.H{})
			return
		}
		if instruction.From == "" || instruction.To == "" || amount <= 0 {
			context.JSON(400, gin.H{})
			return
		}
		if instruction.Mint != "" && instruction.Decimals <= 0 {
			context.JSON(400, gin.H{})
			return
		}

		dbAddress := models.Address{
			Addr: instruction.From,
		}
		err = models.GetOneAddress(&dbAddress, dbAddress.Addr, "SOL")
		if err != nil {
			context.JSON(400, gin.H{})
			return
		}
		if dbAddress.PrivateKey == "" {
			context.JSON(400, gin.H{})
			return
		}

		var fromAccount, e = types.AccountFromBase58(dbAddress.PrivateKey)

		if e != nil {
			context.JSON(400, gin.H{})
			return
		}

		fromAccounts[instruction.From] = &fromAccount
		if instruction.Mint == "" {
			Instructions = append(Instructions, system.Transfer(system.TransferParam{
				From:   fromAccount.PublicKey,
				To:     common.PublicKeyFromString(instruction.To),
				Amount: amount,
			}))
		} else {
			mintPublicKey := common.PublicKeyFromString(instruction.Mint)
			toAddress := common.PublicKeyFromString(instruction.To)
			ataFrom := common.PublicKeyFromString(instruction.AtaFrom)
			var ataTo common.PublicKey
			if instruction.AtaTo == "" {
				ataTo, _, err = common.FindAssociatedTokenAddress(toAddress, mintPublicKey)
				if err != nil { // 没有创建
					context.JSON(400, gin.H{})
					log.Fatal("get_ataTo_account, err: ", toAddress, mintPublicKey.String(), err)
					return
				}
				Instructions = append(Instructions, associated_token_account.Create(associated_token_account.CreateParam{
					Funder:                 fromAccount.PublicKey,
					Owner:                  toAddress,
					Mint:                   mintPublicKey,
					AssociatedTokenAccount: ataTo,
				}))
			} else {
				ataTo = common.PublicKeyFromString(instruction.AtaTo)
			}
			Instructions = append(Instructions, token.TransferChecked(token.TransferCheckedParam{
				From:     ataFrom,
				To:       ataTo,
				Mint:     mintPublicKey,
				Auth:     fromAccount.PublicKey,
				Signers:  []common.PublicKey{},
				Amount:   amount,
				Decimals: instruction.Decimals,
			}))
		}
	}

	if fromAccounts[params.FeePayer] == nil {
		dbAddress := models.Address{
			Addr: params.FeePayer,
		}
		err := models.GetOneAddress(&dbAddress, dbAddress.Addr, "SOL")
		if err != nil {
			context.JSON(400, gin.H{})
			return
		}
		var fromAccount, e = types.AccountFromBase58(dbAddress.PrivateKey)

		if e != nil {
			context.JSON(400, gin.H{})
			return
		}
		fromAccounts[params.FeePayer] = &fromAccount
		return
	}
	feeAccount := fromAccounts[params.FeePayer]

	var signers []types.Account
	for _, value := range fromAccounts {
		signers = append(signers, *value)
	}
	tx, err := types.NewTransaction(types.NewTransactionParam{
		Message: types.NewMessage(types.NewMessageParam{
			FeePayer:        feeAccount.PublicKey,
			RecentBlockhash: params.Blockhash,
			Instructions:    Instructions,
		}),
		Signers: signers,
	})
	if err != nil {
		context.JSON(400, gin.H{})
		log.Println("sign_tx_error, err:", err)
		return
	}
	rawTx, err := tx.Serialize()
	if err != nil {
		context.JSON(400, gin.H{})
		log.Println("failed to serialize tx, err:", err)
		return
	}

	hash := base58.Encode(tx.Signatures[0])
	context.JSON(200, gin.H{
		"hash":  hash,
		"rawTx": base64.StdEncoding.EncodeToString(rawTx),
	})

	//c := client.NewClient(rpc.DevnetRPCEndpoint)
	//_, err = c.SendTransaction(sContext.Background(), tx)
	//if err != nil {
	//	log.Fatal("failed to send tx, err:", err, hash)
	//}
	//
	//log.Println("send tx success, hash:", hash, params.Blockhash)
}

type SignParams struct {
	Blockhash    string        `json:"blockhash"`
	FeePayer     string        `json:"feePayer"`
	Instructions []Instruction `json:"instructions"`
}

type Instruction struct {
	From     string      `json:"from"`
	To       string      `json:"to"`
	AtaFrom  string      `json:"ataFrom"`
	AtaTo    string      `json:"ataTo"`
	Amount   json.Number `json:"amount"`
	Mint     string      `json:"mint"` // 合约地址
	Decimals uint8       `json:"decimals"`
}
