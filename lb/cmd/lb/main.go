package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"strconv"

	"lb"
)

func main() {
	configFile := ""
	pidFile := ""
	startHTTP := false

	flags := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	flags.StringVar(&configFile, "config", "", "config")
	flags.StringVar(&pidFile, "pid-file", "", "write pid to this file")
	flags.BoolVar(&startHTTP, "start-http", startHTTP, "start the HTTP server")
	flags.Parse(os.Args[1:])

	if configFile == "" {
		log.Fatal("error: no config specified")
	}

	if pidFile != "" {
		if err := ioutil.WriteFile(pidFile, []byte(strconv.Itoa(os.Getpid())), 0600); err != nil {
			log.Fatal(err)
		}
		log.Printf("pid [%v] written to %v", os.Getpid(), pidFile)
	}

	m := lb.NewManager(configFile)
	if startHTTP {
		go m.HttpServer()
	}
	m.Run()
	m.Wait()
	log.Println("terminating")
}
