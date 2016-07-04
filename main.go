package main

import (
	"fmt"
	"github.com/FidelityInternational/etcd-leader-monitor/bosh"
	webs "github.com/FidelityInternational/etcd-leader-monitor/web_server"
	"net/http"
	"os"
)

func main() {
	boshConfig := &bosh.Config{
		Username:          os.Getenv("BOSH_USERNAME"),
		Password:          os.Getenv("BOSH_PASSWORD"),
		BoshURI:           os.Getenv("BOSH_URI"),
		Port:              os.Getenv("BOSH_PORT"),
		SkipSSLValidation: true,
	}
	boshClient := bosh.NewClient(boshConfig)

	server := webs.CreateServer(boshClient, &http.Client{})

	router := server.Start()

	http.Handle("/", router)

	err := http.ListenAndServe(":"+os.Getenv("PORT"), nil)
	if err != nil {
		fmt.Println("ListenAndServe:", err)
	}
}
