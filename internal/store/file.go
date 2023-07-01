package store

import (
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/jayakrishnan-jayu/go-sync-env/internal/models"
)

type fileStore struct {
	envStore  *envFileStore
	userStore *userFileStore
}

func (f *fileStore) Env() EnvStore {
	return f.envStore
}

func (f *fileStore) User() UserStore {
	return f.userStore
}

func NewFileStore(envPath string, userPath string) Store {
	var err error
	err = createFileIfNotExists(envPath)
	if err != nil {
		log.Fatalf("Failed to open %s file. %s", envPath, err.Error())
	}
	err = createFileIfNotExists(userPath)
	if err != nil {
		log.Fatalf("Failed to open %s file. %s", envPath, err.Error())
	}
	return &fileStore{
		envStore: &envFileStore{filePath: envPath},
		userStore: &userFileStore{filepath: userPath},
	}
}

type envFileStore struct {
	mutex    sync.RWMutex
	filePath string
}

func (e *envFileStore) Read() ([]byte, error) {
	e.mutex.RLock()
	defer e.mutex.RUnlock()
	return os.ReadFile(e.filePath)
}

func (e *envFileStore) Write(data []byte) (error) {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	return os.WriteFile(e.filePath, data, 0640)
}

type userFileStore struct {
	mutex sync.RWMutex
	filepath string
}

func (u *userFileStore) Users() ([]models.User, error) {
	u.mutex.RLock()
	fileContent, err := os.ReadFile(u.filepath)
	if err != nil {
		return nil, err
	}
	u.mutex.RUnlock()
	fileLines := strings.Split(string(fileContent), "\n")
	users := make([]models.User, 0, len(fileLines))
	for _, line := range fileLines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue 
		}
		fields := strings.Split(line, " ")
		if len(fields) != 3 {
			return nil, errors.New("Invalid user format")
		}

		access, err := strconv.Atoi(fields[2])
		if err != nil || !models.ValidAccessLevel(int64(access)){
			return nil, errors.New("Invalid user access")
		}
		name, err := decode(fields[0])
		if err != nil {
			return nil, err
		}
		key, err := decode(fields[1])
		if err != nil {
			return nil, err
		}
		user := models.User{
			Name:       name,
			PublicKey: key,
			Access: models.AccessLvl(access),
		}
		if !user.Valid() {
			
			return nil, errors.New("invalid user")
		}
		users = append(users, user)
	}
	finalUsers := make([]models.User, len(users))
	copy(finalUsers, users)
	return finalUsers, nil
}

func (u *userFileStore) Add(user models.User) error {
	if !user.Valid() {
		return errors.New("Invalid User")
	}
	u.mutex.Lock()
	defer u.mutex.Unlock()
	file, err := os.OpenFile(u.filepath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(fmt.Sprintf("%s %s %d\n", encode(user.Name), encode(user.PublicKey), user.Access))
	if err != nil {
		return err
	}
	return nil
}

func (u *userFileStore) RemoveByPublicKey(key string) error {
	users, err := u.Users()
	userIndex := -1
	if err != nil {
		return err
	}
	if len(users) == 0 {
		return errors.New("No users found")
	}
	var builder strings.Builder
	for i, user := range users {
		if user.PublicKey == encode(key) {
			userIndex = i
			continue
		}
		builder.WriteString(fmt.Sprintf("%s %s %d\n", encode(user.Name), encode(user.PublicKey), user.Access))
	}
	if userIndex < 0 {
		return errors.New("User not found")
	}

	u.mutex.Lock()
	defer u.mutex.Unlock()
	return os.WriteFile(u.filepath, []byte(builder.String()), 0640)
}

func encode(inp string) string {
	return base64.StdEncoding.EncodeToString([]byte(inp))
}

func decode(inp string) (string, error) {
	decodedBytes, err := base64.StdEncoding.DecodeString(inp)
	if err != nil {
		return "", fmt.Errorf("Error decoding base64: %v", err)
	}
	return string(decodedBytes), nil
}

func createFileIfNotExists(filePath string) error {
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0640)
	if err != nil && !os.IsExist(err) {
		return err
	}
	defer file.Close()
	return nil
}
