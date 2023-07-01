package server

import "os"

type Server interface {

}

type ServerRunner interface {
	Run(signals <-chan os.Signal) error
}