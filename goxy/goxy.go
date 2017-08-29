package goxy

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"regexp"
	"time"
)

type config struct {
	Port string
}

type handler struct {
}

func (handler *handler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	fmt.Println("\na new request-----")
	fmt.Println("method:", req.Method)
	fmt.Println("host:", req.Host)
	fmt.Println("userAgent:", req.UserAgent())
	fmt.Println("body:", req.GetBody)
	fmt.Println("url:", req.URL)
	fmt.Println("header:", req.Header)
	var respRealServer *http.Response
	var connRealServer net.Conn
	var err error
	host := req.Host
	// respClient, _, err := resp.(http.Hijacker).Hijack()

	matched, _ := regexp.MatchString(":[0-9]+$", host)
	if !matched {
		host += ":80"
	}

	connRealServer, err = net.DialTimeout("tcp", host, time.Second*30)
	if err != nil {
		fmt.Println("connRealServer conn err", err)
		return
	}

	err = req.Write(connRealServer)
	if err != nil {
		fmt.Println("connRealServer write err", err)
		return
	}

	respRealServer, err = http.ReadResponse(bufio.NewReader(connRealServer), req)
	if err != nil {
		fmt.Println("respResalServer read err", err)
		return
	}

	if respRealServer == nil {
		fmt.Println("respRealServer is nil")
		return
	}

	respDump, dumpErr := httputil.DumpResponse(respRealServer, true)
	if dumpErr != nil {
		fmt.Println("respResalServer dump err", dumpErr)
		return
	}

	lenRes, er := resp.Write(respDump)
	if er != nil {
		fmt.Println(er)
		return
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
