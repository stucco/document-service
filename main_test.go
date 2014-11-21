package main

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

func NewDocumentClient(addr, base string) *DocumentClient {
	return &DocumentClient{Address: addr, BasePath: base}
}

func (d *DocumentClient) getDoc(id string) (*ResponseType, error) {
	res, err := http.Get(d.Address + "/" + d.BasePath + "/" + id)
	if err != nil {
		return nil, err
	}
	// fmt.Println(res.Status)
	// fmt.Println(res.Header)
	// fmt.Println(res.Body)
	r, err := parseResponse(res)
	if err != nil {
		return nil, err
	}
	return r, nil
}

func (d *DocumentClient) deleteDoc(id string) (*ResponseType, error) {
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

func (d *DocumentClient) postDoc(id, contentType, content string) (*ResponseType, error) {
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

func startServer(port, db string) (*os.Process, error) {
	cmd := exec.Command("go", "run", "main.go", "-port="+port, "-db-file="+db)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	err = cmd.Start()
	// read from stdout and wait to see the port number in the output,
	// which is the last thing the server reports
	wait := 100 * time.Millisecond
	waitedFor := 0 * time.Millisecond
	maxWait := 30 * time.Second
	for {
		so, err := ioutil.ReadAll(stdout)
		if err != nil {
			return nil, err
		}
		if bytes.Contains(so, []byte("on :"+port)) || waitedFor >= maxWait {
			break
		}
		time.Sleep(wait)
		waitedFor += wait
	}
	return cmd.Process, err
}

func stopServer(proc *os.Process) error {
	err := proc.Kill()
	return err
}

func prepare(p int, dbPath string, t *testing.T) (*DocumentClient, *os.Process) {
	port := fmt.Sprintf("%d", p)
	proc, err := startServer(port, dbPath)
	if err != nil {
		t.Errorf("Unable to start server: %s", err)
	}
	c := NewDocumentClient("http://127.0.0.1:"+port, "document")
	return c, proc
}

func teardown(proc *os.Process, dbPath string, t *testing.T) {
	err := stopServer(proc)
	if err != nil {
		t.Errorf("Unable to stop server: %s", err)
	}
	os.Remove(dbPath)
}

func TestJsonUpload(t *testing.T) {
	port := 5050
	db := fmt.Sprintf("test-%d.db", port)
	testContent := "{\"key1\": 123, \"key2\": 456}"
	c, proc := prepare(port, db, t)
	res, err := c.postDoc("", "application/json", testContent)
	if err != nil || !res.Ok {
		t.Errorf("Error uploading json: ", err)
	}
	k := res.Key
	res, err = c.deleteDoc(k)
	if err != nil || !res.Ok {
		t.Errorf("Error deleting json: ", err)
	}
	teardown(proc, db, t)
}

func TestJsonUploadWithKey(t *testing.T) {
	port := 5051
	db := fmt.Sprintf("test-%d.db", port)
	testKey := fmt.Sprintf("k-%d", port)
	testContent := "{\"key1\": 123, \"key2\": 456}"
	c, proc := prepare(port, db, t)
	res, err := c.postDoc(testKey, "application/json", testContent)
	if err != nil || !res.Ok {
		t.Errorf("Error uploading json with key: ", err)
	}
	res, err = c.deleteDoc(testKey)
	if err != nil || !res.Ok {
		t.Errorf("Error deleting json: ", err)
	}
	teardown(proc, db, t)
}

// func TestJsonDownload(t *testing.T) {
// 	port := 5052
// 	db := fmt.Sprintf("test-%d.db", port)
// 	testKey := fmt.Sprintf("k-%d", port)
// 	testContent := "{\"key1\": 123, \"key2\": 456}"
// 	c, proc := prepare(port, db, t)
// 	res, err := c.postDoc(testKey, "application/json", testContent)
// 	if err != nil || !res.Ok {
// 		t.Errorf("Error uploading json with key: ", err)
// 	}
// 	res, err = c.getDoc(testKey)
// 	if err != nil || !res.Ok {
// 		t.Errorf("Error downloading json: ", err)
// 	}
// 	res, err = c.deleteDoc(testKey)
// 	if err != nil || !res.Ok {
// 		t.Errorf("Error deleting json: ", err)
// 	}
// 	fmt.Println(res.Document)
// 	// if res.Document != testContent {
// 	// 	t.Errorf("Error downloading json - unexpected content: ", err)
// 	// }
// 	teardown(proc, db, t)
// }

// func TestJsonDownloadNotExist(t *testing.T) {
// 	c, proc := prepare(5053, t)
// 	k := "a-non-existant-key"
// 	res, err := c.getDoc(k)
// 	if err == nil {
// 		t.Errorf("Error downloading nonexistant json, should have errrored")
// 	}
// 	res, err = c.deleteDoc(key)
// 	if err != nil || !res.Ok {
// 		t.Errorf("Error deleting json: ", err)
// 	}
// 	teardown(proc, t)
// }

// func TestJsonDelete(t *testing.T) {
// 	c, proc := prepare(5054, t)
// 	res, err := c.deleteDoc(testKey)
// 	if err != nil || !res.Ok {
// 		t.Errorf("Error deleting json: ", err)
// 	}
// 	teardown(proc, t)
// }
