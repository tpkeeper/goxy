package goxy

import (
	"bytes"
	"fmt"
	"net/http"
	"time"
)

type config struct {
	Port string
}

type handler struct {
}

func (handler *handler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	fmt.Println("\na new request-----")
	fmt.Println(req.Method)
	fmt.Println(req.Host)
	fmt.Println(req.UserAgent())
	fmt.Println(req.GetBody)
	fmt.Println(req.URL)
	fmt.Println(req.Header)
	lenRes, err := resp.Write(bytes.NewBufferString("hello").Bytes())
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("\nresp %d char\n", lenRes)
}

func StartProxy() error {
	config := config{
		Port: "8080",
	}

	handler := &handler{}

	server := http.Server{
		Addr:         ":" + config.Port,
		Handler:      handler,
		ReadTimeout:  1 * time.Hour,
		WriteTimeout: 1 * time.Hour,
	}

	fmt.Print("start proxy at port:", config.Port)
	err := server.ListenAndServe()
	if err != nil {
		return err
	}
	return nil
}
