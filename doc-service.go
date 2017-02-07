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
	"path"
	"time"

	"github.com/boltdb/bolt"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	glog "github.com/labstack/gommon/log"
	"github.com/satori/go.uuid"
	"github.com/tylerb/graceful"
)

// DocMetadata struct for document metadata to save to database.
type DocMetadata struct {
	Timestamp        int64  `json:"timestamp,omitempty"`
	Name             string `json:"name,omitempty"`
	ContentType      string `json:"content-type,omitempty"`
	Extractor        string `json:"extractor,omitempty"`
	Title            string `json:"title,omitempty"`
	CreationDate     string `json:"creation-date,omitempty"`
	ModificationDate string `json:"modification-date,omitempty"`
}

// ResponseType struct to send as json to client.
type ResponseType struct {
	Ok               bool   `json:"ok,string"`
	Key              string `json:"key,omitempty"`
	Message          string `json:"message,omitempty"`
	Error            string `json:"error,omitempty"`
	Document         string `json:"document,omitempty"`
	Timestamp        int64  `json:"timestamp,omitempty"`
	Name             string `json:"name,omitempty"`
	ContentType      string `json:"content-type,omitempty"`
	Extractor        string `json:"extractor,omitempty"`
	Title            string `json:"title,omitempty"`
	CreationDate     string `json:"creation-date,omitempty"`
	ModificationDate string `json:"modification-date,omitempty"`
}

const (
	// HTTP status code - OK
	statusOk = http.StatusOK
	// HTTP status code - StatusInternalServerError
	statusErr = http.StatusInternalServerError
	// HTTP custom error code - FileExistsError
	fileExistsErr = 515
)

const (
	// Default port.
	defaultPort = 8000
	// Default directory to store the databse.
	dataDir = "data"
	// Database file name.
	dbFileName = "doc.db"
)

// Essentially constants.
var (
	// Database bucket to put metadata in.
	dbBucket = []byte("DocMetadata")
	// Database file path.
	dbFilePath = path.Join(dataDir, dbFileName)
)

var (
	// Verbose logging.
	verbose bool
	// Use GZIP compression.
	useGzip bool
	// Database instance.
	db *bolt.DB
)

func main() {
	port := flag.Int("port", defaultPort, "Port to start the server on")
	flag.BoolVar(&verbose, "debug", false, "Show verbose output")
	flag.BoolVar(&useGzip, "gzip", false, "Use gzip compression")
	flag.Parse()

	var e *echo.Echo
	e = EchoEngine(*port)

	err := os.MkdirAll(dataDir, 0777)
	if err != nil {
		log.Fatalf("Unable to create the data directory %s\n", dataDir)
	}
	db = createDb(dbFilePath, dbBucket)
	defer db.Close()

	graceful.ListenAndServe(e.Server, 5*time.Second)
}

// EchoEngine will create the database and http handler.
func EchoEngine(port int) (e *echo.Echo) {
	e = echo.New()

	e.Pre(middleware.RemoveTrailingSlash())

	e.Logger.SetOutput(os.Stderr)
	e.Use(middleware.Logger())
	if verbose {
		e.Logger.SetLevel(glog.INFO)
	} else {
		e.Logger.SetLevel(glog.ERROR)
	}

	if useGzip {
		e.Use(middleware.GzipWithConfig(middleware.GzipConfig{
			Level: 5,
		}))
	}

	e.Use(middleware.Recover())

	docRoutes := e.Group("/document")
	// Get a document by the document id. Contents are returned in the
	// response body.
	// If there is a failure, the HTTP header and JSON response will
	// indicate it.
	docRoutes.GET("/:id", getDoc)
	// Add a new document, with an assigned id. JSON response indicates
	// success or failure.
	docRoutes.POST("", newDoc)
	// Add a new document, passing an id. JSON response indicates success
	// or failure.
	docRoutes.POST("/:id", newDocWithID)
	// Remove a document based on the id. JSON response indicates success
	// or failure.
	docRoutes.DELETE("/:id", deleteDoc)

	e.Server.Addr = fmt.Sprintf(":%d", port)
	e.Server.WriteTimeout = 90 * time.Second
	e.Server.ReadTimeout = 60 * time.Second

	return e
}

// Create and return the bolt database for storing metadata.
func createDb(f string, bucket []byte) *bolt.DB {
	database, err := bolt.Open(f, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.Fatalf("Unable to create the metadata database %s: %s", f, err)
	}
	err = database.Update(func(tx *bolt.Tx) error {
		_, err2 := tx.CreateBucketIfNotExists(bucket)
		if err2 != nil {
			log.Fatalf("Unable to create the metadata database bucket %s: %s", bucket, err2)
		}
		return nil
	})
	if err != nil {
		log.Fatalf("Unable to update the metadata database bucket %s: %s", bucket, err)
	}
	return database
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
		err2 := b.Put([]byte(key), buf.Bytes())
		return err2
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
func getDoc(c echo.Context) error {
	key := c.Param("id")
	filePath := dataDir + "/" + key
	fs, err := os.Stat(filePath)
	if err != nil || fs.Size() <= 0 {
		return c.JSON(statusErr, newErrorResp(key, "key not found", err))
	}
	f, err := os.Open(dataDir + "/" + key)
	defer f.Close()
	if err != nil {
		return c.JSON(statusErr, newErrorResp(key, "unable to open data", err))
	}
	d, err := ioutil.ReadAll(f)
	if err != nil {
		return c.JSON(statusErr, newErrorResp(key, "error reading file", err))
	}

	metadata, err := getMetadata(key)
	if err != nil {
		return c.JSON(statusErr, newErrorResp(key, "error reading metadata", err))
	}
	r := newSuccessResp(key, "")
	r.Timestamp = metadata.Timestamp
	r.Name = metadata.Name
	r.ContentType = metadata.ContentType
	r.Extractor = metadata.Extractor
	r.Title = metadata.Title
	r.CreationDate = metadata.CreationDate
	r.ModificationDate = metadata.ModificationDate
	r.Document = string(d)

	return c.JSON(statusOk, r)
}

// Add a new document, creating a new v4 UUID.
func newDoc(c echo.Context) error {
	key := uuid.NewV4().String()
	res := saveDocument(key, c)
	if res.Ok == false {
		if res.Message == "file exists" {
			return c.JSON(fileExistsErr, res)
		}
		return c.JSON(statusErr, res)
	}
	return c.JSON(statusOk, res)
}

// Add a new document, using the provided id.
func newDocWithID(c echo.Context) error {
	key := c.Param("id")
	res := saveDocument(key, c)
	if res.Ok == false {
		return c.JSON(statusErr, res)
	}
	return c.JSON(statusOk, res)
}

// Delete document from disk and metadata from database.
func deleteDoc(c echo.Context) error {
	key := c.Param("id")
	err := os.Remove(dataDir + "/" + key)
	if err != nil {
		return c.JSON(statusErr, newErrorResp(key, "error removing document", err))
	}
	err = deleteMetadata(key)
	if err != nil {
		return c.JSON(statusErr, newErrorResp(key, "error removing metadata", err))
	}
	return c.JSON(statusOk, newSuccessResp(key, "removed document"))
}

// Save document to disk and metadata to database.
func saveDocument(key string, c echo.Context) *ResponseType {
	body := c.Request().Body
	defer body.Close()
	filePath := dataDir + "/" + key
	fi, err := os.Stat(filePath)
	if err == nil && fi.Size() > 0 {
		return newErrorResp(key, "file exists", fmt.Errorf("file already exists for key %s", key))
	}
	f, err := os.Create(filePath)
	if err != nil {
		return newErrorResp(key, "file creation error", fmt.Errorf("error creating file for key %s: %s", key, err.Error()))
	}
	defer f.Close()
	size, err := io.Copy(f, body)
	if size == 0 {
		return newErrorResp("", "input error", fmt.Errorf("no data uploaded"))
	}
	if err != nil {
		return newErrorResp(key, "file write error", fmt.Errorf("error copying body to file for key %s: %s", key, err.Error()))
	}
	name := c.Request().FormValue("name")
	contentType := c.Request().Header.Get("Content-Type")
	extractor := c.Request().FormValue("extractor")
	title := c.Request().FormValue("dc:title")
	creation := c.Request().FormValue("dcterms:created")
	modification := c.Request().FormValue("dcterms:modified")
	metadata := DocMetadata{
		Timestamp:        time.Now().Unix(),
		Name:             name,
		ContentType:      contentType,
		Extractor:        extractor,
		Title:            title,
		CreationDate:     creation,
		ModificationDate: modification,
	}
	err = saveMetadata(key, &metadata)
	if err != nil {
		return newErrorResp(key, "file metadata write error", fmt.Errorf("error saving metadata for key %s: %s", key, err.Error()))
	}
	return newSuccessResp(key, fmt.Sprintf("document saved (%d bytes)", size))
}

// Create a new error response to send to client.
func newErrorResp(key, msg string, err error) *ResponseType {
	return &ResponseType{Ok: false, Message: msg, Error: err.Error(), Key: key}
}

// Create a new success response to send to client.
func newSuccessResp(key, msg string) *ResponseType {
	return &ResponseType{Ok: true, Message: msg, Key: key}
}
