package main_test

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
)

type DocumentClient struct {
	Address  string
	BasePath string
}

type DocumentResponse struct {
	Ok      string `json:"ok"`
	Key     string `json:"key"`
	Error   string `json:"error"`
	Message string `json:"message"`
	Data    []byte
}

func NewDocumentClient(addr, base string) *DocumentClient {
	return &DocumentClient{Address: addr, BasePath: base}
}

func (d *DocumentClient) getDoc(id string) (*DocumentResponse, error) {
	res, err := http.Get(d.Address + "/" + d.BasePath + "/" + id)
	defer res.Body.Close()
	if err != nil {
		return nil, err
	}
	result := &DocumentResponse{Key: id}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	result.Data = body
	return result, nil
}

func (d *DocumentClient) deleteDoc(id string) (*DocumentResponse, error) {
	client := &http.Client{}
	req, err := http.NewRequest("DELETE", d.Address+"/"+d.BasePath+"/"+id, nil)
	if err != nil {
		return nil, err
	}
	res, err := client.Do(req)
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	var r DocumentResponse
	err = json.Unmarshal(body, &r)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (d *DocumentClient) postDoc(id, contentType, content string) (*DocumentResponse, error) {
	uri := d.Address + "/" + d.BasePath + "/"
	if id != "" {
		uri += id
	}
	buf := bytes.NewBufferString(content)
	res, err := http.Post(uri, contentType, buf)
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	var r DocumentResponse
	err = json.Unmarshal(body, &r)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (d *DocumentClient) postDocFile(id, contentType, filePath string) (*DocumentResponse, error) {
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
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	var r DocumentResponse
	err = json.Unmarshal(body, &r)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func TestJsonUploadWithKey(t *testing.T) {
	c := NewDocumentClient("http://127.0.0.1:8000", "document")
	key := "12345678"
	content := "{'key1': 123, 'key2': 456}"
	res, err := c.postDoc(key, "application/json", content)
	if err != nil {
		t.Errorf("Error uploading json: ", err)
	}
	if res.Ok != "true" {
		t.Errorf("Error uploading json '%s', response: %v", content, res)
	}
}

func TestJsonDownload(t *testing.T) {
	c := NewDocumentClient("http://127.0.0.1:8000", "document")
	key := "12345678"
	expectedContent := "{'key1': 123, 'key2': 456}"
	res, err := c.getDoc(key)
	if err != nil {
		t.Errorf("Error downloading json: ", err)
	}
	if string(res.Data) != expectedContent {
		t.Errorf("Error downloading json: ", err)
	}
}

func TestJsonDelete(t *testing.T) {
	c := NewDocumentClient("http://127.0.0.1:8000", "document")
	key := "12345678"
	_, err := c.deleteDoc(key)
	if err != nil {
		t.Errorf("Error deleting json: ", err)
	}
}
