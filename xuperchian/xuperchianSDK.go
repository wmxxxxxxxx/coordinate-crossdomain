package xuperchian

import (
	"fmt"
	"github.com/xuperchain/xuper-sdk-go/v2/account"
	"github.com/xuperchain/xuper-sdk-go/v2/xuper"
	"io/ioutil"
)

const (
	Address = "RzF1azYaFEoLyH6dLxuz33YUb1njDkRAt"
	Contract_Addr = "XC1234567890111113@xuper"
	Contract_Name = "SDKNativeCount1"
	Mnemonic = "付 说 愿 城 看 无 牲 恨 策 丝 骤 百"
)

// CreateAccount
/*
@Description: 创建账户
	命令行转钱：
	./bin/xchain-cli transfer --to RzF1azYaFEoLyH6dLxuz33YUb1njDkRAt --amount 100000000 --keys data/keys/ -H 127.0.0.1:37101
*/
func CreateAccount() {
	var acc *account.Account
	var err error
	// 测试创建账户
	acc, err = account.CreateAccount(1, 1)
	if err != nil {
		fmt.Printf("create account error: %v\n", err)
	}
	fmt.Println(acc)
	fmt.Println(acc.Mnemonic)
	return
}

// CreateContractAccount
/*
@Description: 创建合约账户
	命令行给合约账户转钱：
	./bin/xchain-cli transfer --to XC1234567890111113@xuper --amount 1000000000000
*/
func CreateContractAccount() {
	// 从文件中恢复账户
	acc, err := account.RetrieveAccount(Mnemonic, 1)
	if err != nil {
		fmt.Printf("retrieveAccount err: %v\n", err)
		return
	}
	fmt.Printf("RetrieveAccount: to %v\n", acc)

	// 创建一个合约账户
	// 合约账户是由 (XC + 16个数字 + @xuper) 组成, 比如 "XC1234567890123456@xuper"
	contractAccount := Contract_Addr

	xchainClient, err := xuper.New("127.0.0.1:37101")
	tx, err := xchainClient.CreateContractAccount(acc, contractAccount)
	if err != nil {
		fmt.Printf("createContractAccount err:%s\n", err.Error())
	}
	fmt.Println(tx.Tx.Txid)
	return
}

func getAccount() *account.Account{
	account, err := account.RetrieveAccount(Mnemonic, 1)
	if err != nil {
		fmt.Printf("retrieveAccount err: %v\n", err)
		return nil
	}
	contractAccount := Contract_Addr
	err = account.SetContractAccount(contractAccount)
	if err != nil {
		panic(err)
	}
	return account
}

func DeployContract() {
	codePath := "contract/contract" // 编译好的二进制文件 go build -o counter
	code, err := ioutil.ReadFile(codePath)
	if err != nil {
		panic(err)
	}
	account := getAccount()
	contractName := Contract_Name
	xchainClient, err := xuper.New("127.0.0.1:37101")
	if err != nil {
		panic(err)
	}
	args := map[string]string{
		"creator": "xuperchain",
		"key":     "contract",
	}
	var tx *xuper.Transaction
	tx, err = xchainClient.DeployNativeGoContract(account, contractName, code, args)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Deploy Native Go Contract Success! %x\n", tx.Tx.Txid)
}

func run()  {
	getAccount()
}