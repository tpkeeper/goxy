package goxy

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"regexp"
	"time"
)

type config struct {
	Port string //监听端口号
	Dump bool   //是否解析请求和返回
}

type handler struct {
	Config config
}

func StartProxy() error {

	handler := &handler{
		Config: config{
			Port: "8080",
			Dump: true,
		},
	}

	server := http.Server{
		Addr:         ":" + handler.Config.Port,
		Handler:      handler,
		ReadTimeout:  1 * time.Hour,
		WriteTimeout: 1 * time.Hour,
	}

	fmt.Print("start proxy at port:", handler.Config.Port)
	err := server.ListenAndServe()
	if err != nil {
		return err
	}
	return nil
}

func (handler *handler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {

	if req.Method == "CONNECT" {
		HttpsProxy(&resp, req)
		return
	}
	HttpProxy(&resp, req, handler.Config)
}

func HttpProxy(respWriter *http.ResponseWriter, req *http.Request, config config) {
	var respRealServer *http.Response
	var connRealServer net.Conn
	var err error

	if config.Dump {
		DumpReq(req)
	}

	req.Header.Del("Proxy-Connection")
	req.Header.Set("Connection", "Keep-Alive")
	host := req.Host
	respClient, _, err := (*respWriter).(http.Hijacker).Hijack()
	if err != nil {
		fmt.Println("hijack to client resp err")
	}
	defer respClient.Close()
	matched, _ := regexp.MatchString(":[0-9]+$", host)
	if !matched {
		host += ":80"
	}

	connRealServer, err = net.DialTimeout("tcp", host, time.Second*30)
	if err != nil {
		fmt.Println("connRealServer conn err", err)
		return
	}
	fmt.Println("connect to real server success")

	err = req.Write(connRealServer)
	if err != nil {
		fmt.Println("connRealServer write err", err)
		return
	}

	fmt.Println("write to real server success")

	respRealServer, err = http.ReadResponse(bufio.NewReader(connRealServer), req)
	if err != nil {
		fmt.Println("respResalServer read err", err)
		return
	}

	if respRealServer == nil {
		fmt.Println("RealServerl have no response")
		return
	}

	if config.Dump {
		DumpResp(respRealServer)
	}

	respDump, dumpErr := httputil.DumpResponse(respRealServer, true)
	if dumpErr != nil {
		fmt.Println("respResalServer dump err", dumpErr)
		return
	}

	lenRes, er := respClient.Write(respDump)
	if er != nil {
		fmt.Println(er)
		return
	}
	fmt.Println("write to client success")
	fmt.Printf("\nresp %d char\n", lenRes)
}

func HttpsProxy(resp *http.ResponseWriter, req *http.Request) {
	var connRealServer net.Conn
	respClient, _, err := (*resp).(http.Hijacker).Hijack()
	if err != nil {
		fmt.Println("hijack to client resp err")
	}
	connRealServer, err = net.DialTimeout("tcp", req.Host, time.Second*30)
	if err != nil {
		fmt.Println("connRealServer conn err", err)
		return
	}
	_, er := respClient.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
	if er != nil {
		fmt.Println(er)
		return
	}
	Transport(connRealServer, respClient)
}

//两个io口的连接
func Transport(conn1, conn2 net.Conn) (err error) {
	rChan := make(chan error, 1)
	wChan := make(chan error, 1)

	go MyCopy(conn1, conn2, wChan)
	go MyCopy(conn2, conn1, rChan)
	select {
	case err = <-wChan:
	case err = <-rChan:
	}
	return
}
func MyCopy(src io.Reader, dst io.Writer, ch chan<- error) {
	_, err := io.Copy(dst, src)
	ch <- err
}
