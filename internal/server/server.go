package server

import (
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	"github.com/gliderlabs/ssh"
	"github.com/jayakrishnan-jayu/go-sync-env/internal/cache"
	"github.com/jayakrishnan-jayu/go-sync-env/internal/models"
	"github.com/jayakrishnan-jayu/go-sync-env/internal/store"
	gossh "golang.org/x/crypto/ssh"
)

type ServerConfig struct {
	SSHPort    int
	
}

func DefaultServerConfig() ServerConfig {
	return ServerConfig{
		SSHPort: 2222,
	}
}

type server struct {
	store store.Store
	cache cache.Cache
	sshPort int
	srv *ssh.Server
}

func NewServer(config *ServerConfig, store store.Store, privateKey []byte) (ServerRunner, error) {
	signer, err := gossh.ParsePrivateKey(privateKey)
	if err != nil {
		return nil, err
	}

	users, err := store.User().Users()
	if err != nil {
		return nil, err
	}
	serverCache := cache.NewCache(users)

	s := &server{
		store: store,
		cache: serverCache,
		sshPort: config.SSHPort,
	}


	srv := &ssh.Server{
		ConnectionFailedCallback: func(conn net.Conn, err error) {
			log.Println("ssh connection failed, ", conn.RemoteAddr(), conn.LocalAddr(), err)
		},
		Addr: fmt.Sprintf("0.0.0.0:%d", config.SSHPort),
		Handler:     s.handler,
		HostSigners: []ssh.Signer{signer},
		PublicKeyHandler: func(ctx ssh.Context, key ssh.PublicKey) bool {
			keyString := strings.TrimRight(string(gossh.MarshalAuthorizedKey(key)), "\n") 
			exists := s.cache.User().ExistsByPublicKey(keyString)
			log.Println(keyString)
			// if exists {
			// 	user, ok := s.cache.User().ByPublicKey(keyString)
			// 	if !ok {
			// 		log.Println("faild to get user ", err)
			// 	}
			// }
			return exists
		},
		ServerConfigCallback: func(ctx ssh.Context) *gossh.ServerConfig {
			return &gossh.ServerConfig{
				NoClientAuth: false,
			}
		},
	}

	s.srv = srv


	return serverRunner{server: s, listenAddr: fmt.Sprintf("0.0.0.0:%d", s.sshPort)}, nil
}

func (s *server) handler(session ssh.Session) {
	publicKey := session.PublicKey()
	if publicKey == nil {
		_ = session.Exit(1)
		return
	}
	keyString := strings.TrimRight(string(gossh.MarshalAuthorizedKey(publicKey)), "\n") 
	user, ok := s.cache.User().ByPublicKey(keyString)
	if !ok {
		log.Println("Couldn't not find user in cache ")
		_ = session.Exit(1)
		return
	}

	commands := session.Command()
	if len(commands) == 0 || len(commands) > 1 {
		session.Write([]byte("Invalid args\n"))
		_ = session.Exit(1)
		return
	}
	cmd := commands[0]
	buffer := make([]byte, 1024)
	switch cmd {
	case "push":
		// Perform actions for the "push" command
		// ...
		if user.Access < models.WriteAccess {
			log.Println("user does not have push access")
			session.Write([]byte("You do not have push access\n"))
			_ = session.Exit(1)
			return
		}
		n, err := session.Read(buffer)
		if err != nil {
			log.Printf("Failed to read data from client: %v", err)
			break
		}
		data := buffer[:n]
		// TODO : data validation
		err = s.store.Env().Write(data)
		if err != nil {
			log.Println("faild to to write data", data)
			_ = session.Exit(1)
			return
		}
	case "pull":
		if user.Access < models.ReadAccess {
			log.Println("user does not have pull access")
			session.Write([]byte("You do not have pull access\n"))
			_ = session.Exit(1)
			return
		}
		serverFileBytes, err := s.store.Env().Read()
		if err != nil {
			log.Println(err)
			_ = session.Exit(1)
			return
		}
		_, err = session.Write(serverFileBytes)
		if err != nil {
			log.Println(err)
			_ = session.Exit(1)
			return
		}
	default:
		session.Write([]byte("Unknown command\n"))
		_ = session.Exit(1)
		return
	}

	_ = session.Exit(0)
}


type serverRunner struct {
	server *server
	listenAddr string
}


func (s serverRunner) Run(signals <-chan os.Signal) error {
	
	exited := make(chan struct{})


	go func() {
		defer close(exited)
		// s.server.srv.Serve(listener)
		err := s.server.srv.ListenAndServe()
		log.Println("Stoped listening", err)
	}()

	for {
		select {
		case <-exited:
			return nil
		case <-signals:
			s.server.srv.Close()
			// listener.Close()
		}
	}
}

