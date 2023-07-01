package models

import (
	"fmt"

	"golang.org/x/crypto/ssh"
)

type AccessLvl int64

const (
	NoAccess AccessLvl = iota
	ReadAccess
	WriteAccess
)

type User struct {
	Name string
	PublicKey string
	Access AccessLvl 
}

func (u *User) Valid() bool {
	return u.Name != "" && 
	isValidSSHPublicKey(u.PublicKey) &&
	ValidAccessLevel(int64(u.Access))
}

func (u *User) String() string {
	return fmt.Sprintf("%s %s %d", u.Name, u.PublicKey, u.Access)
}

func ValidAccessLevel(i int64) bool {
	return i >= 0 && i <= int64(WriteAccess)
}

func isValidSSHPublicKey(publicKeyStr string) bool {
	_, _, _, _, err := ssh.ParseAuthorizedKey([]byte(publicKeyStr))
	return err == nil
}