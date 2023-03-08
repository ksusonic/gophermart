package api

type User struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type BalanceResponse struct {
	Current   int64 `json:"current"`
	Withdrawn int64 `json:"withdrawn"`
}
