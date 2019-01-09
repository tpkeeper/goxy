package goxy

import (
	"net/http"
	"fmt"
	"compress/gzip"
	"io/ioutil"
	"compress/flate"
	"bytes"
	"bufio"
	"strings"
	"io"
)

/**
解析http请求，此处传递的是Request的一个拷贝
 */
func DumpReq(r *http.Request) {
	fmt.Println("\nRequest:")
	fmt.Printf("%s %s %s\n", r.Method, r.Host, r.RequestURI)
	fmt.Printf("%s %s \n", "RemoteAddr", r.RemoteAddr)
	for headerName, headerValue := range r.Header {
		fmt.Printf("%s : %s \n", headerName, headerValue)
	}
	err := r.ParseForm()
	if err != nil {
		fmt.Println(err)
	}
	if len(r.Form) > 0 {
		if r.Method == "POST" {
			fmt.Println("POST PARAM:")
		}
		if r.Method == "GET" {
			fmt.Println("GET PARAM:")
		}
		for key, value := range r.Form {
			fmt.Printf("%s : %s\n", key, value)
		}
	}

}

/**
解析服务端返回数据，传入指针不会出现Body close错误，未知原因
 */
func DumpResp(res *http.Response) {
	needDumpBody := false
	fmt.Println("\nResponse:")
	for key, value := range res.Header {
		fmt.Printf("%s : %s \n", key, value)
	}
	contentType := res.Header["Content-Type"]
	fmt.Println(contentType)
	for _, value := range contentType {
		fmt.Println("value", value)
		switch {
		case strings.Contains(value, "text/html"):
			needDumpBody = true
		case strings.Contains(value, "text/plain"):
			needDumpBody = true
		case strings.Contains(value, "application/json"):
			needDumpBody = true
		}
	}
	if !needDumpBody {
		return
	}
	save := res.Body
	var err error
	save, res.Body, err = drainBody(res.Body) //copy body

	contentEncodings := res.Header["Content-Encoding"]
	bytesBody, err := ioutil.ReadAll(res.Body)
	//bytesBody, err := ioutil.ReadAll(save)
	if err != nil {
		fmt.Println("dumpResponse: res.Body read err:", err)
		return
	}

	res.Body = save // recover body

	var bytesBuffer bytes.Buffer
	writer := bufio.NewWriter(&bytesBuffer)
	writer.Write(bytesBody)
	writer.Flush()

	for _, value := range contentEncodings {
		switch value {
		case "gzip":
			reader, err := gzip.NewReader(&bytesBuffer)
			if err != nil {
				fmt.Println(err)
			} else {
				defer reader.Close()
				bytesBody, err = ioutil.ReadAll(reader)
				if err != nil {
					fmt.Println(err)
					return
				}
			}
			fmt.Println("ggggggggggggggg")
			break
		case "deflate":
			reader := flate.NewReader(&bytesBuffer)
			defer reader.Close()
			bytesBody, err = ioutil.ReadAll(reader)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println("dddddddddddddd")
			break
		}
	}

	fmt.Println("Body:")
	fmt.Println(string(bytesBody))
}

// drainBody reads all of b to memory and then returns two equivalent
// ReadClosers yielding the same bytes.
//
// It returns an error if the initial slurp of all bytes fails. It does not attempt
// to make the returned ReadClosers have identical error-matching behavior.
func drainBody(b io.ReadCloser) (r1, r2 io.ReadCloser, err error) {
	if b == http.NoBody {
		// No copying needed. Preserve the magic sentinel meaning of NoBody.
		return http.NoBody, http.NoBody, nil
	}
	var buf bytes.Buffer
	if _, err = buf.ReadFrom(b); err != nil {
		return nil, b, err
	}
	if err = b.Close(); err != nil {
		return nil, b, err
	}
	return ioutil.NopCloser(&buf), ioutil.NopCloser(bytes.NewReader(buf.Bytes())), nil
}
