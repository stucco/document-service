package main_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"testing"
	"time"
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

func startServer(port string) (*os.Process, error) {
	cmd := exec.Command("go", "run", "main.go", "-port="+port)
	err := cmd.Start()
	return cmd.Process, err
}

func stopServer(proc *os.Process) error {
	err := proc.Kill()
	return err
}

func prepare(p int, t *testing.T) (*DocumentClient, *os.Process) {
	port := fmt.Sprintf("%d", p)
	proc, err := startServer(port)
	time.Sleep(2500)
	if err != nil {
		t.Errorf("Unable to start server: %s", err)
	}
	c := NewDocumentClient("http://127.0.0.1:"+port, "document")
	return c, proc
}

func teardown(proc *os.Process, t *testing.T) {
	err := stopServer(proc)
	if err != nil {
		t.Errorf("Unable to stop server: %s", err)
	}
}

func TestJsonUploadWithKey(t *testing.T) {
	c, proc := prepare(5051, t)
	key := "12345678"
	content := "{'key1': 123, 'key2': 456}"
	res, err := c.postDoc(key, "application/json", content)
	if err != nil {
		t.Errorf("Error uploading json: ", err)
	}
	if res.Ok != "true" {
		t.Errorf("Error uploading json '%s', response: %v", content, res)
	}
	teardown(proc, t)
}

func TestJsonDownload(t *testing.T) {
	c, proc := prepare(5052, t)
	key := "12345678"
	expectedContent := "{'key1': 123, 'key2': 456}"
	res, err := c.getDoc(key)
	if err != nil {
		t.Errorf("Error downloading json: ", err)
	}
	if string(res.Data) != expectedContent {
		t.Errorf("Error downloading json: ", err)
	}
	teardown(proc, t)
}

func TestJsonDelete(t *testing.T) {
	c, proc := prepare(5053, t)
	key := "12345678"
	_, err := c.deleteDoc(key)
	if err != nil {
		t.Errorf("Error deleting json: ", err)
	}
	teardown(proc, t)
}
