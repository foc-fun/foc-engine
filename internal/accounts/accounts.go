package accounts

import "fmt"

type AccountInfo struct {
	Username string `json:"username"`
}

type Accounts struct {
	AccountsContractAddress string
	// Map: <contract address>::<account address> -> <account info>
	AccountsInfo map[string]AccountInfo
}

var FocAccounts *Accounts

func AddAccountsContract(address string) {
	FocAccounts = &Accounts{
		AccountsContractAddress: address,
	}
}

func GetAccountsContract() string {
	if FocAccounts == nil {
		return ""
	}
	return FocAccounts.AccountsContractAddress
}

func GetFocAccountInfo(accountAddress string) *AccountInfo {
	if FocAccounts == nil {
		return nil
	}
	// Default method of storing account is on the FocAccounts contract
	accountsContract := FocAccounts.AccountsContractAddress
	accountKey := fmt.Sprintf("%s::%s", accountsContract, accountAddress)
	if info, ok := FocAccounts.AccountsInfo[accountKey]; ok {
		return &info
	}
	return nil
}

func SetAccountsInfo(contractAddress string, accountAddress string, info *AccountInfo) {
	if FocAccounts == nil {
		return
	}
	// Default method of storing account is on the FocAccounts contract
	accountKey := fmt.Sprintf("%s::%s", contractAddress, accountAddress)
	if info != nil {
		FocAccounts.AccountsInfo[accountKey] = *info
	} else {
		delete(FocAccounts.AccountsInfo, accountKey)
	}
}

func GetContractAccountInfo(contractAddress string, accountAddress string) *AccountInfo {
	// TODO: Load from mongo?
	// TODO: Do proper zero padding for addresses
	if FocAccounts == nil {
		return nil
	}
	// Default method of storing account is on the FocAccounts contract
	accountKey := fmt.Sprintf("%s::%s", contractAddress, accountAddress)
	if info, ok := FocAccounts.AccountsInfo[accountKey]; ok {
		return &info
	}
	return nil
}
