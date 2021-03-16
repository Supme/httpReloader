# HTTP Reloader

[![Go Report Card](https://goreportcard.com/badge/github.com/Supme/httpreloader)](https://goreportcard.com/report/github.com/Supme/httpreloader)
[![GoDoc](https://godoc.org/github.com/Supme/httpreloader?status.svg)](https://pkg.go.dev/github.com/Supme/httpreloader?tab=doc)

## Features

* Update certificate without restart http server
* Multidomain with support for multiple domain and wildcard domain certificates, the default certificate for non-specified domains.

## Examples

``` golang
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
```

more examples in example_test.go or see code 