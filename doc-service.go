// This software provides a storage service for text documents and metadata over an HTTP API.

package main

import (
	"bytes"
	"encoding/gob"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/stucco/document-service/Godeps/_workspace/src/code.google.com/p/go-uuid/uuid"
	"github.com/stucco/document-service/Godeps/_workspace/src/github.com/boltdb/bolt"
	"github.com/stucco/document-service/Godeps/_workspace/src/github.com/gin-gonic/gin"
)

// Document metadata to save to database.
type DocMetadata struct {
	Timestamp   int64  `json:"timestamp,omitempty"`
	Name        string `json:"name,omitempty"`
	ContentType string `json:"content-type,omitempty"`
	Extractor   string `json:"extractor,omitempty"`
}

// Response struct to send as json to client.
type ResponseType struct {
	Ok          bool   `json:"ok,string"`
	Key         string `json:"key,omitempty"`
	Message     string `json:"message,omitempty"`
	Error       string `json:"error,omitempty"`
	Document    string `json:"document,omitempty"`
	Timestamp   int64  `json:"timestamp,omitempty"`
	Name        string `json:"name,omitempty"`
	ContentType string `json:"content-type,omitempty"`
	Extractor   string `json:"extractor,omitempty"`
}

const (
	// HTTP status code - OK
	statusOk = http.StatusOK
	// HTTP status code - StatusInternalServerError
	statusErr = http.StatusInternalServerError
)

var (
	// Database instance.
	db *bolt.DB
	// Database bucket to put metadata in.
	dbBucket []byte
	// Relative or absolute path to the directory to save documents in.
	dataDir string
)

func main() {

	flag.StringVar(&dataDir, "doc-dir", "./data", "Directory to store documents")
	port := flag.Int("port", 8000, "Port to start the server on")
	debug := flag.Bool("debug", false, "Show debug output")
	flag.Parse()

	dbBucket = []byte("DocMetadata")

	runtime.GOMAXPROCS(runtime.NumCPU())

	err := os.MkdirAll(dataDir, 0777)
	if err != nil {
		log.Fatalf("Unable to create the data directory %s\n", dataDir)
	}

	dbFile := dataDir + "/doc.db"
	db = createDb(&dbFile, &dbBucket)
	defer db.Close()

	if !*debug {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())

	docs := r.Group("/document")
	{
		// Get a document by the document id. Contents are returned in the response body.
		// If there is a failure, the HTTP header and JSON response will indicate it.
		docs.GET("/:id", getDoc)
		// Add a new document, passing an id. JSON response indicates success or failure.
		docs.POST("/:id", newDocWithId)
		// Add a new document, with an assigned id. JSON response indicates success or failure.
		docs.POST("/", newDoc)
		// Remove a document based on the id. JSON response indicates success or failure.
		docs.DELETE("/:id", deleteDoc)
	}

	if !*debug {
		fmt.Printf("Listening and serving on :%d", *port)
	}
	r.Run(fmt.Sprintf(":%d", *port))
}

// Create and return the bolt database for storing metadata.
func createDb(f *string, bucket *[]byte) *bolt.DB {
	db, err := bolt.Open(*f, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.Fatalf("Unable to create the metadata database %s: %s", *f, err)
	}
	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(*bucket)
		if err != nil {
			log.Fatalf("Unable to create the metadata database bucket %s: %s", *bucket, err)
		}
		return nil
	})
	return db
}

// Add metadata to the database.
func saveMetadata(key string, metadata *DocMetadata) error {
	buf := &bytes.Buffer{}
	enc := gob.NewEncoder(buf)
	err := enc.Encode(metadata)
	if err != nil {
		return err
	}
	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(dbBucket)
		err := b.Put([]byte(key), buf.Bytes())
		return err
	})
	return err
}

// Get metadata based on an id.
func getMetadata(id string) (*DocMetadata, error) {
	var metadata DocMetadata
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(dbBucket)
		v := b.Get([]byte(id))
		buf := bytes.NewBuffer(v)
		dec := gob.NewDecoder(buf)
		err := dec.Decode(&metadata)
		return err
	})
	return &metadata, err
}

// Delete metadata based on an id.
func deleteMetadata(id string) error {
	err := db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(dbBucket).Delete([]byte(id))
	})
	return err
}

// Get a document and metadata.
func getDoc(c *gin.Context) {
	key := c.Params.ByName("id")
	filePath := dataDir + "/" + key
	fs, err := os.Stat(filePath)
	if err != nil || fs.Size() <= 0 {
		c.JSON(statusErr, newErrorResp(key, "key not found", err))
		return
	}
	f, err := os.Open(dataDir + "/" + key)
	defer f.Close()
	if err != nil {
		c.JSON(statusErr, newErrorResp(key, "unable to open data", err))
		return
	}
	d, err := ioutil.ReadAll(f)
	if err != nil {
		c.JSON(statusErr, newErrorResp(key, "error reading file", err))
		return
	}

	metadata, err := getMetadata(key)
	if err != nil {
		c.JSON(statusErr, newErrorResp(key, "error reading metadata", err))
		return
	}
	r := newSuccessResp(key, "")
	r.Timestamp = metadata.Timestamp
	r.Name = metadata.Name
	r.ContentType = metadata.ContentType
	r.Extractor = metadata.Extractor
	r.Document = string(d)
	c.JSON(statusOk, r)
}

// Add a new document, creating a new v4 UUID.
func newDoc(c *gin.Context) {
	key := uuid.New()
	err := saveDocument(key, c)
	if err != nil {
		c.JSON(statusErr, newErrorResp(key, "error saving document", err))
	} else {
		c.JSON(statusOk, newSuccessResp(key, "saved document"))
	}
}

// Add a new document, using the provided id.
func newDocWithId(c *gin.Context) {
	key := c.Params.ByName("id")
	err := saveDocument(key, c)
	if err != nil {
		c.JSON(statusErr, newErrorResp(key, "error saving document", err))
	} else {
		c.JSON(statusOk, newSuccessResp(key, "saved document by id"))
	}
}

// Delete document from disk and metadata from database.
func deleteDoc(c *gin.Context) {
	key := c.Params.ByName("id")
	err := os.Remove(dataDir + "/" + key)
	if err != nil {
		c.JSON(statusErr, newErrorResp(key, "error removing document", err))
	} else {
		err = deleteMetadata(key)
		if err != nil {
			c.JSON(statusErr, newErrorResp(key, "error removing metadata", err))
		} else {
			c.JSON(statusOk, newSuccessResp(key, "removed document"))
		}
	}
}

// Save document to disk and metadata to database.
func saveDocument(key string, c *gin.Context) error {
	filePath := dataDir + "/" + key
	fi, err := os.Stat(filePath)
	if err == nil && fi.Size() > 0 {
		return fmt.Errorf("file already exists for key %s", key)
	}
	f, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("error creating file for key %s: %s", key, err.Error())
	}
	defer f.Close()
	_, err = io.Copy(f, c.Request.Body)
	if err != nil {
		return fmt.Errorf("error copying body to file for key %s: %s", key, err.Error())
	}
	name := c.Request.FormValue("name")
	contentType := c.Request.Header.Get("Content-Type")
	fmt.Println(contentType)
	fmt.Println(c.Request.Header)
	extractor := c.Request.FormValue("extractor")
	metadata := DocMetadata{Timestamp: time.Now().Unix(), Name: name, ContentType: contentType, Extractor: extractor}
	err = saveMetadata(key, &metadata)
	if err != nil {
		return fmt.Errorf("error saving metadata for key %s: %s", key, err.Error())
	}
	return nil
}

// Create a new error response to send to client.
func newErrorResp(key, msg string, err error) *ResponseType {
	return &ResponseType{Ok: false, Message: msg, Error: err.Error(), Key: key}
}

// Create a new success response to send to client.
func newSuccessResp(key, msg string) *ResponseType {
	return &ResponseType{Ok: true, Message: msg, Key: key}
}
