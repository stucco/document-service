package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"
)

const (
	port = "8267"
)

type DocClient struct {
	Address  string
	BasePath string
}

func NewDocClient(addr, base string) *DocClient {
	return &DocClient{Address: addr, BasePath: base}
}

func (d *DocClient) getDoc(id string) (*ResponseType, error) {
	res, err := http.Get(d.Address + "/" + d.BasePath + "/" + id)
	if err != nil {
		return nil, err
	}
	r, err := parseResponse(res)
	if err != nil {
		return nil, err
	}
	return r, nil
}

func (d *DocClient) deleteDoc(id string) (*ResponseType, error) {
	client := &http.Client{}
	req, err := http.NewRequest("DELETE", d.Address+"/"+d.BasePath+"/"+id, nil)
	if err != nil {
		return nil, err
	}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	r, err := parseResponse(res)
	if err != nil {
		return nil, err
	}
	return r, nil
}

func (d *DocClient) postDoc(id, contentType, content string) (*ResponseType, error) {
	uri := d.Address + "/" + d.BasePath + "/"
	if id != "" {
		uri += id
	}
	buf := bytes.NewBufferString(content)
	res, err := http.Post(uri, contentType, buf)
	if err != nil {
		return nil, err
	}
	r, err := parseResponse(res)
	if err != nil {
		return nil, err
	}
	return r, nil
}

func parseResponse(res *http.Response) (*ResponseType, error) {
	body, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		return nil, err
	}
	var r ResponseType
	err = json.Unmarshal(body, &r)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func TestJSONUpload(t *testing.T) {
	testContent := "{\"k1\": 123, \"k2\": 456}"
	c := NewDocClient("http://127.0.0.1:"+port, "document")
	res, err := c.postDoc("", "application/json", testContent)
	if err != nil || !res.Ok {
		t.Errorf("Error uploading json: %s", err.Error())
	}
	k := res.Key
	res, err = c.deleteDoc(k)
	if err != nil || !res.Ok {
		t.Errorf("Error deleting json: %s", err.Error())
	}
}

// func TestJSONUploadWithKey(t *testing.T) {
// 	testKey := "2345-6789"
// 	testContent := "{\"k1\": 123, \"k2\": 456}"
// 	c := NewDocClient("http://127.0.0.1:"+port, "document")
// 	res, err := c.postDoc(testKey, "application/json", testContent)
// 	if err != nil || !res.Ok {
// 		t.Errorf("Error uploading json with key: %s", err.Error())
// 	}
// 	res, err = c.deleteDoc(testKey)
// 	if err != nil || !res.Ok {
// 		t.Errorf("Error deleting json: %s", err.Error())
// 	}
// }
//
// func TestJSONDownload(t *testing.T) {
// 	testKey := "2345-6789"
// 	testContent := "{\"k1\": 123, \"k2\": 456}"
// 	c := NewDocClient("http://127.0.0.1:"+port, "document")
// 	res, err := c.postDoc(testKey, "application/json", testContent)
// 	if err != nil || !res.Ok {
// 		t.Errorf("Error uploading json with key: %s", err.Error())
// 	}
// 	res, err = c.getDoc(testKey)
// 	if err != nil || !res.Ok {
// 		t.Errorf("Error downloading json: %s", err.Error())
// 	}
// 	if res.Document != testContent {
// 		t.Errorf("Error downloading json - unexpected content for document with key %s:\n %s", testKey, err.Error())
// 	}
// 	res, err = c.deleteDoc(testKey)
// 	if err != nil || !res.Ok {
// 		t.Errorf("Error deleting json: %s", err.Error())
// 	}
// }
