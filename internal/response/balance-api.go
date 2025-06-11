package response

type BalancesResponse struct {
	Balances []TokenBalance `json:"balances"`
}

type TokenBalance struct {
	TokenPath string `json:"tokenPath"`
	Amount    int64  `json:"amount"`
}

type AccountBalance struct {
	Address   string `json:"address"`
	TokenPath string `json:"tokenPath"`
	Amount    int64  `json:"amount"`
}

type AccountBalancesResponse struct {
	AccountBalances []AccountBalance `json:"accountBalances"`
}
