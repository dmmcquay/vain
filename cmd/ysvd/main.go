package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"mcquay.me/vain"

	"github.com/kelseyhightower/envconfig"
)

const usage = `ysvd

environment vars:

YSV_PORT: tcp listen port
`

type config struct {
	Port int
}

func main() {
	c := &config{
		Port: 4040,
	}
	if err := envconfig.Process("ysv", c); err != nil {
		fmt.Fprintf(os.Stderr, "problem processing environment: %v", err)
		os.Exit(1)
	}
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "env", "e", "help", "h":
			fmt.Fprintf(os.Stderr, "%s\n", usage)
			os.Exit(1)
		}
	}
	hostname := "localhost"
	if hn, err := os.Hostname(); err != nil {
		log.Printf("problem getting hostname:", err)
	} else {
		hostname = hn
	}
	log.Printf("serving at: http://%s:%d/", hostname, c.Port)
	sm := http.NewServeMux()
	vain.NewServer(sm)
	addr := fmt.Sprintf(":%d", c.Port)
	if err := http.ListenAndServe(addr, sm); err != nil {
		log.Printf("problem with http server: %v", err)
		os.Exit(1)
	}
}
