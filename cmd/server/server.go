package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/jayakrishnan-jayu/go-sync-env/internal/server"
	"github.com/jayakrishnan-jayu/go-sync-env/internal/store"
)

func main() {

	// This private key was generated for test purpose and is okay to be included in the version control
	serverPrivateKey := `-----BEGIN OPENSSH PRIVATE KEY-----
b3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAAAMwAAAAtzc2gtZW
QyNTUxOQAAACCweXXXBO2tJuyaXhVET0g+ZLwIbirJBj+olU9BsRJQrAAAALCsA2RhrANk
YQAAAAtzc2gtZWQyNTUxOQAAACCweXXXBO2tJuyaXhVET0g+ZLwIbirJBj+olU9BsRJQrA
AAAEDClHHJuhEJvwbUPBGvHNNPr8VKdIfRMbNONKy91m2HkLB5ddcE7a0m7JpeFURPSD5k
vAhuKskGP6iVT0GxElCsAAAALGpheWFrcmlzaG5hbkBKYXlha3Jpc2huYW5zLU1hY0Jvb2
stQWlyLmxvY2FsAQ==
-----END OPENSSH PRIVATE KEY-----
`

	fileStore := store.NewFileStore(".hidden/.env.server", ".hidden/.user.server")

	// fileStore.User().Add(models.User{Name: "user", PublicKey: "public key", Access: 2})
	serverConfig := server.DefaultServerConfig()
	server, err := server.NewServer(&serverConfig, fileStore, []byte(serverPrivateKey))
	if err != nil {
		log.Fatal(err)
	}

	signals := make(chan os.Signal)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	done := make(chan struct{})
	
	go func() {
		if err := server.Run(signals); err != nil {
			log.Printf("Server exited with error: %s\n", err)
			os.Exit(1)
		}
		close(done)

	}()

	<-done
	log.Println("Exiting...")
	
}