package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"

	"github.com/stucco/document-service/Godeps/_workspace/src/code.google.com/p/go-uuid/uuid"
	"github.com/stucco/document-service/Godeps/_workspace/src/github.com/gin-gonic/gin"
)

const (
	dataDir = "./data"
)

func main() {

	port := flag.Int("port", 8000, "Port to start the server on")
	debug := flag.Bool("debug", false, "Show debug output")
	flag.Parse()

	runtime.GOMAXPROCS(runtime.NumCPU())

	if !*debug {
		gin.SetMode(gin.ReleaseMode)
	}

	err := os.MkdirAll(dataDir, 0777)
	if err != nil {
		fmt.Printf("Unable to create the data directory %s\n", dataDir)
		os.Exit(1)
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

	// r.NotFound(func() (int, string) {
	// 	return 404, "Route was not found"
	// })

	if !*debug {
		fmt.Printf("Listening and serving on :%d", *port)
	}
	r.Run(fmt.Sprintf(":%d", *port))
}

func getDoc(c *gin.Context) {
	key := c.Params.ByName("id")
	filePath := dataDir + "/" + key
	_, err := os.Stat(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": "false", "error": "key " + key + " not found."})
		return
	}
	// c.File(filePath)

	f, err := os.Open(dataDir + "/" + key)
	defer f.Close()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": "false", "error": "key " + key + " not found."})
		return
	}
	_, err = io.Copy(c.Writer, f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": "false", "error": "error copying file to response for " + key + ": " + err.Error()})
	}
}

func newDoc(c *gin.Context) {
	key := uuid.New()
	errCode, err := addDocument(key, c)
	if err != nil {
		c.JSON(errCode, gin.H{"ok": "false", "error": err.Error()})
	} else {
		c.JSON(http.StatusOK, gin.H{"ok": "true", "key": key, "message": "document added"})
	}
}

func newDocWithId(c *gin.Context) {
	key := c.Params.ByName("id")
	errCode, err := addDocument(key, c)
	if err != nil {
		c.JSON(errCode, gin.H{"ok": "false", "error": err.Error()})
	} else {
		c.JSON(http.StatusOK, gin.H{"ok": "true", "key": key, "message": "document added with id"})
	}
}

func deleteDoc(c *gin.Context) {
	key := c.Params.ByName("id")
	err := os.Remove(dataDir + "/" + key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": "false", "error": "cannot remove file " + key + ":" + err.Error()})
	} else {
		c.JSON(http.StatusOK, gin.H{"ok": "true", "key": key, "message": "document removed"})
	}
}

func addDocument(key string, c *gin.Context) (int, error) {
	filePath := dataDir + "/" + key
	fi, err := os.Stat(filePath)
	if err == nil && fi.Size() > 0 {
		return http.StatusInternalServerError, fmt.Errorf("file already exists for key %s", key)
	}
	f, err := os.Create(filePath)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("error creating file for key %s: %s", key, err.Error())
	}
	defer f.Close()
	_, err = io.Copy(f, c.Request.Body)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("error copying body to file for key %s: %s", key, err.Error())
	}
	return 0, nil
}
