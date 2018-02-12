package main

import (
	"encoding/json"
	"log"
	"os"
	"time"
)

type VCapApp struct {
	ApplicationName string `json:"application_name"`
}

func main() {
	vcap := os.Getenv("VCAP_APPLICATION")
	if vcap == "" {
		log.Fatalf("missing required environment variable VCAP_APPLICATION")
	}

	var vcapApp VCapApp
	err := json.Unmarshal([]byte(vcap), &vcapApp)
	if err != nil {
		log.Fatalf("failed to unmarshal VCAP_APPLICATION")
	}

	for {
		log.Printf("APP_LOG: %s", vcapApp.ApplicationName)
		time.Sleep(50 * time.Millisecond)
	}
}
