package blockchain

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

func main() {
	url := "http://localhost:2006/contract/invoke"
	method := "POST"
	payload := strings.NewReader(`{` + " " + ` "contractAddress" : "0xc860ab27901b3c2b810165a6096c64d88763617f",` + " " + ` "functionName" : "storeEvidence",` + " " + ` "args" : ["3","touteng"],` + " " + ` "memberName" :"pcm",` + " " + ` "type": "2"` + " " + ` }`)
	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Add("User-Agent", "Apifox/1.0.0 (https://apifox.com)")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "localhost:2006")
	req.Header.Add("Connection", "keep-alive")
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(body))
}
