package main

import (
	"fmt"
	webs "github.com/FidelityInternational/etcd-leader-monitor/web_server"
	"github.com/cloudfoundry-community/gogobosh"
	"net/http"
	"os"
)

func main() {
	boshConfig := &gogobosh.Config{
		Username:          os.Getenv("BOSH_USERNAME"),
		Password:          os.Getenv("BOSH_PASSWORD"),
		BOSHAddress:       os.Getenv("BOSH_URI"),
		SkipSslValidation: true,
	}
	boshClient, err := gogobosh.NewClient(boshConfig)
	if err != nil {
		fmt.Println("Could not create bosh client")
		fmt.Println(err)
		os.Exit(1)
	}

	server := webs.CreateServer(boshClient, &http.Client{})

	router := server.Start()

	http.Handle("/", router)

	err = http.ListenAndServe(":"+os.Getenv("PORT"), nil)
	if err != nil {
		fmt.Println("ListenAndServe:", err)
	}
}
