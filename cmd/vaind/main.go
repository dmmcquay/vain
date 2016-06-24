package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"mcquay.me/vain"

	"github.com/kelseyhightower/envconfig"
)

const usage = "vaind <dbname>"

type config struct {
	Port     int
	Insecure bool

	Cert string
	Key  string

	Static string

	EmailTimeout time.Duration `envconfig:"email_timeout"`

	SMTPHost string `envconfig:"smtp_host"`
	SMTPPort int    `envconfig:"smtp_port"`

	From string
}

func main() {
	log.SetFlags(log.Lshortfile)
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "%s\n", usage)
		os.Exit(1)
	}

	c := &config{
		Port:         4040,
		EmailTimeout: 5 * time.Minute,
		SMTPPort:     25,
	}
	if err := envconfig.Process("vain", c); err != nil {
		fmt.Fprintf(os.Stderr, "problem processing environment: %v", err)
		os.Exit(1)
	}
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "env", "e":
			fmt.Printf("VAIN_PORT:           %v\n", c.Port)
			fmt.Printf("VAIN_INSECURE:       %v\n", c.Insecure)
			fmt.Printf("VAIN_CERT:           %v\n", c.Cert)
			fmt.Printf("VAIN_KEY:            %v\n", c.Key)
			fmt.Printf("VAIN_STATIC:         %v\n", c.Static)
			fmt.Printf("VAIN_EMAIL_TIMEOUT:  %v\n", c.EmailTimeout)
			fmt.Printf("VAIN_SMTP_HOST:      %v\n", c.SMTPHost)
			fmt.Printf("VAIN_SMTP_PORT:      %v\n", c.SMTPPort)
			fmt.Printf("VAIN_FROM:           %v\n", c.From)
			os.Exit(0)
		case "help", "h":
			fmt.Printf("%s\n", usage)
			os.Exit(0)
		}
	}
	log.Printf("%+v", c)

	db, err := vain.NewMemDB(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "couldn't open db: %v\n", err)
		os.Exit(1)
	}

	sigs := make(chan os.Signal)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
	go func() {
		s := <-sigs
		log.Printf("signal: %+v", s)
		if err := db.Sync(); err != nil {
			log.Printf("problem syncing db to disk: %+v", err)
			os.Exit(1)
		}
		os.Exit(0)
	}()

	m, err := vain.NewMail(c.From, c.SMTPHost, c.SMTPPort)
	if err != nil {
		fmt.Fprintf(os.Stderr, "problem initializing mailer: %v", err)
		os.Exit(1)
	}

	hostname := "localhost"
	if hn, err := os.Hostname(); err != nil {
		log.Printf("problem getting hostname: %v", err)
	} else {
		hostname = hn
	}
	log.Printf("serving at: http://%s:%d/", hostname, c.Port)
	sm := http.NewServeMux()
	vain.NewServer(sm, db, m, c.Static, c.EmailTimeout, c.Insecure)
	addr := fmt.Sprintf(":%d", c.Port)

	if c.Cert == "" || c.Key == "" {
		log.Printf("INSECURE MODE")
		if err := http.ListenAndServe(addr, sm); err != nil {
			log.Printf("problem with http server: %v", err)
			os.Exit(1)
		}
	} else {
		if err := http.ListenAndServeTLS(addr, c.Cert, c.Key, sm); err != nil {
			log.Printf("problem with http server: %v", err)
			os.Exit(1)
		}
	}
}
