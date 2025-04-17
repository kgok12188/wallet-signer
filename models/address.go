package models

import (
	_ "github.com/go-sql-driver/mysql"
	"wallet-signer/config"
)

type Address struct {
	Addr       string `gorm:"Column:addr" json:"addr"`
	PrivateKey string `gorm:"Column:private_key" json:"privateKey"`
	ChainId    string `gorm:"Column:chain_id" json:"chainId"`
}

func (Address) TableName() string {
	return "address"
}

func GetOneAddress(ret *Address, address, chainId string) (err error) {
	if err := config.DB.Where("addr = ? and chain_id = ?", address, chainId).First(ret).Error; err != nil {
		return err
	}
	return nil
}

func AddNewAddress(address *Address) (err error) {
	if err = config.DB.Create(address).Error; err != nil {
		return err
	}
	return nil
}
