package main

import (
	"flag"
	"fmt"
	"log"

	"gopkg.in/mgo.v2"
	"gopkg.in/tylerb/graceful.v1"

	"github.com/VoycerAG/gridfs-image-server/server"
)

// main starts the server and returns an invalid result as exit code
func main() {
	configurationFilepath := flag.String("config", "configuration.json", "path to the configuration file")
	serverPort := flag.Int("port", 8000, "the server port where we will serve images")
	host := flag.String("host", "localhost:27017", "the database host with an optional port, localhost would suffice")
	newrelicKey := flag.String("license", "", "your newrelic license key in order to enable monitoring")

	flag.Parse()

	if *configurationFilepath == "" {
		log.Fatal("configuration must be given")
		return
	}

	config, err := server.NewConfigFromFile(*configurationFilepath)
	if err != nil {
		log.Fatal(err)
		return
	}

	session, err := mgo.Dial(*host)
	if err != nil {
		log.Fatal(err)
		return
	}

	session.SetSyncTimeout(0)
	session.SetMode(mgo.Eventual, true)

	storage, err := server.NewGridfsStorage(session)
	if err != nil {
		log.Fatal(err)
		return
	}

	imageServer := server.NewImageServerWithNewRelic(config, storage, *newrelicKey)

	handler := imageServer.Handler()

	log.Printf("Server started. Listening on %d database host is %s\n", *serverPort, *host)

	graceful.Run(fmt.Sprintf(":%d", *serverPort), 0, handler)
	if err != nil {
		log.Fatal(err)
	}
}
