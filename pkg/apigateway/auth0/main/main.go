package main

import (
	"log"
	"net/http"

	"github.com/joho/godotenv"

	"github.com/GoogleCloudPlatform/spark-on-k8s-operator/pkg/apigateway/auth0/authenticator"
	"github.com/GoogleCloudPlatform/spark-on-k8s-operator/pkg/apigateway/auth0/router"
)

// How to run:
// 1. Set work directory to be: /Your/path/datapunchorg/spark-on-k8s-operator/pkg/apigateway/auth0
// 2. Add .env file in work directory:
//AUTH0_DOMAIN=''
//AUTH0_CLIENT_ID=''
//AUTH0_CLIENT_SECRET=''
//AUTH0_CALLBACK_URL=''

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Failed to load the env vars: %v", err)
	}

	auth, err := authenticator.New()
	if err != nil {
		log.Fatalf("Failed to initialize the authenticator: %v", err)
	}

	rtr := router.New(auth)

	log.Print("Server listening on http://localhost:3000/")
	if err := http.ListenAndServe("0.0.0.0:3000", rtr); err != nil {
		log.Fatalf("There was an error with the http server: %v", err)
	}
}
