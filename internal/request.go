package internal

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

func initRequest(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Sailabove sailgo CLI/"+VERSION)
}

func getHTTPClient() *http.Client {
	tr := &http.Transport{}
	return &http.Client{Transport: tr}
}

// GetWantJSON GET on path and return string of JSON
func GetWantJSON(path string) string {
	return GetJSON(ReqWant("GET", http.StatusOK, path, nil))
}

// PostWantJSON POST on path and return string of JSON
func PostWantJSON(path string) string {
	return GetJSON(ReqWant("POST", http.StatusCreated, path, nil))
}

// PostBodyWantJSON POST a body on path and return string of JSON
func PostBodyWantJSON(path string, body []byte) string {
	return GetJSON(ReqWant("POST", http.StatusCreated, path, body))
}

// DeleteWantJSON on path and return string of JSON
func DeleteWantJSON(path string) string {
	return GetJSON(ReqWant("DELETE", http.StatusOK, path, nil))
}

// DeleteBodyWantJSON on path and return string of JSON
func DeleteBodyWantJSON(path string, body []byte) string {
	return GetJSON(ReqWant("DELETE", http.StatusOK, path, body))
}

// ReqWantJSON requests with a method on a path, check wantCode and returns string of JSON
func ReqWantJSON(method string, wantCode int, path string, body []byte) string {
	return GetJSON(ReqWant(method, wantCode, path, body))
}

// StreamWant request a path with method and stream result
func StreamWant(method string, wantCode int, path string, jsonStr []byte) {
	apiRequest(method, wantCode, path, jsonStr, true)
}

// ReqWant requests with a method on a path, check wantCode and returns []byte
func ReqWant(method string, wantCode int, path string, jsonStr []byte) []byte {
	return apiRequest(method, wantCode, path, jsonStr, false)
}

func apiRequest(method string, wantCode int, path string, jsonStr []byte, stream bool) []byte {

	err := ReadConfig()
	if err != nil {
		fmt.Printf("Error reading configuration: %s\n", err)
		os.Exit(1)
	}

	var req *http.Request
	if jsonStr != nil {
		req, _ = http.NewRequest(method, Host+path, bytes.NewReader(jsonStr))
	} else {
		req, _ = http.NewRequest(method, Host+path, nil)
	}

	initRequest(req)
	req.SetBasicAuth(User, Password)
	resp, err := getHTTPClient().Do(req)
	Check(err)
	defer resp.Body.Close()

	var body []byte
	if !stream {
		body, err = ioutil.ReadAll(resp.Body)
	}

	if resp.StatusCode != wantCode || Verbose {
		fmt.Printf("Response Status : %s\n", resp.Status)
		fmt.Printf("Request path : %s\n", Host+path)
		fmt.Printf("Request Headers : %s\n", req.Header)
		fmt.Printf("Request Body : %s\n", string(jsonStr))
		fmt.Printf("Response Headers : %s\n", resp.Header)
		if err == nil {
			fmt.Printf("Response Body : %s\n", string(body))
		}
		if !Verbose {
			os.Exit(1)
		}
	}

	if stream {
		reader := bufio.NewReader(resp.Body)
		for {
			line, err := reader.ReadBytes('\n')
			if err != nil {
				return nil
			}
			if string(line) != "" {
				log.Print(string(line))
			}
		}
	} else {
		Check(err)
		return body
	}
}

// Request executes an authentificated HTTP request on $path given $method and $args
func Request(method string, path string, args []byte) ([]byte, int, error) {
	respBody, code, err := Stream(method, path, args)
	if err != nil {
		return nil, 0, err
	}

	if respBody == nil {
		panic("what ?")
	}
	defer respBody.Close()

	var body []byte
	body, err = ioutil.ReadAll(respBody)
	if err != nil {
		return nil, 0, err
	}

	if Verbose {
		fmt.Printf("Response Body : %s\n", body)
	}

	return body, code, nil
}

// Stream makes an authenticated http request and return io.ReadCloser
func Stream(method string, path string, args []byte) (io.ReadCloser, int, error) {
	err := ReadConfig()
	if err != nil {
		fmt.Printf("Error reading configuration: %s\n", err)
		os.Exit(1)
	}

	var req *http.Request
	if args != nil {
		req, _ = http.NewRequest(method, Host+path, bytes.NewReader(args))
	} else {
		req, _ = http.NewRequest(method, Host+path, nil)
	}

	initRequest(req)
	req.SetBasicAuth(User, Password)
	resp, err := getHTTPClient().Do(req)
	if err != nil {
		return nil, 0, err
	}

	if Verbose {
		fmt.Printf("Response Status : %s\n", resp.Status)
		fmt.Printf("Request path : %s\n", Host+path)
		fmt.Printf("Request Headers : %s\n", req.Header)
		fmt.Printf("Request Body : %s\n", string(args))
		fmt.Printf("Response Headers : %s\n", resp.Header)
	}

	return resp.Body, resp.StatusCode, nil
}

// DisplayStream decode each line from http buffer and print either message or error
func DisplayStream(buffer io.ReadCloser) error {
	reader := bufio.NewReader(buffer)

	for {
		line, err := reader.ReadBytes('\n')
		m := DecodeMessage(line)
		if m != nil {
			fmt.Println(m.Message)
		}
		e := DecodeError(line)
		if e != nil {
			fmt.Println(e)
		}
		if err != nil && err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
	}
}

// GetListApplications returns list of applications, GET on /applications
func GetListApplications(args []string) []string {
	apps := []string{}
	if len(args) == 0 {
		b := ReqWant("GET", http.StatusOK, "/applications", nil)
		err := json.Unmarshal(b, &apps)
		Check(err)
	}
	return apps
}

// GetJSON return string of JSON, prettify if flag pretty is true
func GetJSON(s []byte) string {
	if Pretty {
		var out bytes.Buffer
		json.Indent(&out, s, "", "  ")
		return out.String()
	}
	return string(s)
}

// Check checks e and panic if not nil
func Check(e error) {
	if e != nil {
		panic(e)
	}
}
