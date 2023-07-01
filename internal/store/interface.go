package store

import "github.com/jayakrishnan-jayu/go-sync-env/internal/models"


type Store interface {
	Env() EnvStore
	User() UserStore
}

type EnvStore interface {
	Read() ([]byte, error)
	Write([]byte) error
}

type UserStore interface {
	Users() ([]models.User, error)
	Add(models.User) error
	RemoveByPublicKey(string) error
}

