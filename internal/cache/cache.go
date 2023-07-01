package cache

import (
	"sync"

	"github.com/jayakrishnan-jayu/go-sync-env/internal/models"
)

type cache struct {
	user *userCache
}

func NewCache(users []models.User) Cache {
	cacheUsers := make([]models.User, len(users))
	copy(cacheUsers, users)
	uc := userCache{
		users: cacheUsers,
	}
	return &cache{
		user: &uc,
	}
}

func (c *cache) User() UserCache {
	return c.user 
}


type userCache struct {
	sync.RWMutex
	users []models.User
}

func (u *userCache) ExistsByPublicKey(key string) bool {
	u.RLock()
	defer u.RUnlock()
	for _, user := range u.users {
		if user.PublicKey == key {
			return true
		}
	}
	return false
}

func (u *userCache) ByPublicKey(key string) (models.User, bool) {
	u.RLock()
	defer u.RUnlock()
	for _, user := range u.users {
		if user.PublicKey == user.PublicKey {
			return user, true
		}
	}
	return models.User{}, false
}

func (u *userCache) Update(newUsers []models.User) {
	u.Lock()
	defer u.Unlock()
	u.users = make([]models.User, len(newUsers))
	copy(u.users, newUsers)
}