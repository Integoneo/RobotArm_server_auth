package domain

type User struct {
	UUID         string
	Email        string
	PasswordHash string
	Name         string
	Surname      string
}
