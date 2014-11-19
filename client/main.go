package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

type DocumentClient struct {
	Address  string
	BasePath string
}

type DocumentResponse struct {
	Ok      string `json:"ok"`
	Key     string `json:"key"`
	Message string `json:"message"`
}

type DocumentResult struct {
	Key  string
	Data []byte
}

func NewDocumentClient(addr, base string) *DocumentClient {
	return &DocumentClient{Address: addr, BasePath: base}
}

func (d *DocumentClient) getDoc(id string) (*DocumentResult, error) {
	res, err := http.Get(d.Address + "/" + d.BasePath + "/" + id)
	defer res.Body.Close()
	if err != nil {
		return nil, err
	}
	result := &DocumentResult{Key: id}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	result.Data = body
	return result, nil
}

func (d *DocumentClient) postDoc(id, contentType, filePath string) (*DocumentResult, error) {
	uri := d.Address + "/" + d.BasePath + "/"
	if id != "" {
		uri += id
	}
	f, err := os.Open(filePath)
	defer f.Close()
	if err != nil {
		return nil, err
	}
	res, err := http.Post(uri, contentType, f)
	if err != nil {
		return nil, err
	}
	result := &DocumentResult{}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	var r DocumentResponse
	err = json.Unmarshal(body, &r)
	if err != nil {
		return nil, err
	}
	result.Key = r.Key
	return result, nil
}

func main() {
	host := flag.String("host", "127.0.0.1", "Endpoint host")
	port := flag.Int("port", 8000, "Endpoint port")
	method := flag.String("method", "get", "HTTP method (get, post, delete)")
	key := flag.String("key", "", "Document key to get, post, delete")
	inFile := flag.String("in-file", "", "File to send as input into post")
	inType := flag.String("content-type", "", "Content type of the file to post")
	out := flag.String("out", "", "File to save the output of a get into (if not specified, output will be printed to the console")
	flag.Parse()

	c := NewDocumentClient(fmt.Sprintf("http://%s:%d", *host, *port), "document")

	m := strings.ToLower(*method)
	switch m {
	case "get":
		res, err := c.getDoc(*key)
		if err != nil {
			log.Fatalln(err)
		}
		fmt.Printf("id: %s\n", res.Key)
		if *out == "" {
			fmt.Printf("data: %s\n", res.Data)
		} else {
			fmt.Printf("output: %s\n", *out)
			ioutil.WriteFile(*out, res.Data, 0644)
		}
	case "post":
		if *inFile == "" {
			log.Fatalln("Specify a file to upload.")
		}
		if *inType == "" {
			log.Fatalln("Specify the content-type of the file to upload.")
		}
		res, err := c.postDoc(*key, *inType, *inFile)
		if err != nil {
			log.Fatalln(err)
		}
		fmt.Printf("id: %s\n", res.Key)
	default:
		log.Fatalln("Unknown method")
	}

}
