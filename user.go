package app

// goweb:Model
type User struct {
	ID    uint   `json:"id"`    // goweb:select
	Email string `json:"email"` // goweb:select,insert
}
