package cache

import "github.com/jayakrishnan-jayu/go-sync-env/internal/models"

type Cache interface {
	User() UserCache
}

type UserCache interface {
	Update([]models.User)
	ExistsByPublicKey(string) bool
	ByPublicKey(string) (models.User, bool)
}