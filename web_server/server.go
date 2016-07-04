package webServer

import (
	"github.com/FidelityInternational/etcd-leader-monitor/bosh"
	"github.com/gorilla/mux"
	"net/http"
)

// Server struct
type Server struct {
	Controller *Controller
}

// CreateServer - creates a server
func CreateServer(boshClient *bosh.Client, etcdHTTPClient *http.Client) *Server {
	controller := CreateController(boshClient, etcdHTTPClient)

	return &Server{
		Controller: controller,
	}
}

// Start - starts the web server
func (s *Server) Start() *mux.Router {
	router := mux.NewRouter()

	router.HandleFunc("/", s.Controller.CheckLeaders).Methods("GET")

	return router
}
