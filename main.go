package main

import (
	"github.com/gin-gonic/gin"
	"wallet-signer/chains/eth"
	"wallet-signer/chains/tron"
)

func main() {
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	r.POST("/eth/getAddress", eth.GetAddress)
	r.POST("/eth/sign", eth.SignTx)
	r.POST("/tron/getAddress", tron.GetAddress)
	r.POST("/tron/sign", tron.SignTx)
	err := r.Run(":8080")
	if err != nil {
		panic(err)
	}

}
