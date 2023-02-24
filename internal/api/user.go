package api

type User struct {
	Login    string `json:"login"`
	Password string `json:"password"`

	Current   float64 `json:"current,omitempty"`
	Withdrawn int64   `json:"withdrawn,omitempty"`
}
