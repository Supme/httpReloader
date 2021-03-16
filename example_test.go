package httpreloader

// openssl req -newkey rsa:2048 \
//  -new -nodes -x509 \
//  -days 3650 \
//  -out cert.pem \
//  -keyout key.pem \
//  -subj "/C=US/ST=California/L=Mountain View/O=Your Organization/OU=Your Unit/CN=localhost"

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func ExampleNewServer() {
	addr := ":4443"
	certFile := "./test_data/cert1.pem"
	keyFile := "./test_data/key1.pem"

	server, err := NewServer(addr, certFile, keyFile, nil)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGHUP)
		for range c {
			err := server.Reloader.UpdateCertificate(certFile, keyFile)
			if err != nil {
				log.Print(err)
			}
		}
	}()

	fmt.Println("Listen on:", addr)
	log.Fatal(server.ListenAndServeTLS())
}

func ExampleNewServerReloadWithNewCert() {
	var addr = ":4443"
	server, err := NewServer(addr, "./test_data/cert1.pem", "./test_data/key1.pem", nil)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		s := 1
		timer := time.NewTimer(5 * time.Second)
		for {
			<-timer.C
			s++
			if s > 2 {
				s = 1
			}
			fmt.Println("Loading certificate number", s)
			err = server.Reloader.UpdateCertificate(fmt.Sprintf("./test_data/cert%d.pem", s), fmt.Sprintf("./test_data/key%d.pem", s))
			if err != nil {
				fmt.Println(err)
			}
		}
	}()

	fmt.Println("Listen on:", addr)
	log.Fatal(server.ListenAndServeTLS())
}
