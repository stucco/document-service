package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"testing"

	"github.com/appleboy/gofight"
	"github.com/boltdb/bolt"
	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
)

const (
	testPort      = 8123
	testKey       = "123"
	testTitle     = "test-title"
	testExtractor = "test-extractor"
)

var (
	engine   *echo.Echo
	testJSON = `{
		"a": 1,
		"b": 2
	}`
)

func TestMain(m *testing.M) {
	// Setup before tests.
	dataDirExists := false
	tmpDataDir := "tempDataDir"
	// If there is an existing directory, move it out of the way.
	if _, err := os.Stat(dataDir); err == nil {
		fmt.Printf("Moving original data dir '%s'...\n", dataDir)
		os.Rename(dataDir, tmpDataDir)
		dataDirExists = true
	}

	// Setup the router.
	engine = EchoEngine(testPort)

	// Setup the database.
	err := os.MkdirAll(dataDir, 0777)
	if err != nil {
		log.Fatalf("Unable to create the data directory %s\n", dataDir)
	}
	db = createDb(dbFilePath, dbBucket)
	defer db.Close()
	fmt.Printf("database created '%s'\n", dbFilePath)

	// Run the tests.
	retCode := m.Run()

	// Teardown after tests.
	fmt.Printf("Cleaning up test data dir '%s'...\n", dataDir)
	if err := os.RemoveAll(dataDir); err != nil {
		fmt.Printf("Unable to cleanup directory '%s': %s\n", dataDir, err.Error())
		os.Exit(1)
	}
	if dataDirExists {
		fmt.Printf("Moving original data dir '%s' back...\n", dataDir)
		os.Rename(tmpDataDir, dataDir)
	}

	os.Exit(retCode)
}

func TestPostJSONDoc(t *testing.T) {
	r := gofight.New()
	r.POST("/document").
		SetBody(testJSON).
		SetDebug(true).
		Run(engine, func(r gofight.HTTPResponse, rq gofight.HTTPRequest) {
			assert.Equal(t, http.StatusOK, r.Code)
			data := []byte(r.Body.String())
			var resp ResponseType
			err := json.Unmarshal(data, &resp)
			if assert.NoError(t, err) {
				assert.True(t, resp.Ok, "Response ok should be true")
			}
			cleanupDoc(t, resp.Key)
		})
}

// The document created here is used for testing GET and DELETE.
func TestPostJSONDocWithID(t *testing.T) {
	r := gofight.New()
	r.POST("/document/"+testKey).
		SetQuery(gofight.H{
			"extractor": testExtractor,
			"dc:title":  testTitle,
		}).
		SetBody(testJSON).
		SetDebug(true).
		Run(engine, func(r gofight.HTTPResponse, rq gofight.HTTPRequest) {
			assert.Equal(t, http.StatusOK, r.Code)
			data := []byte(r.Body.String())
			var resp ResponseType
			err := json.Unmarshal(data, &resp)
			if assert.NoError(t, err) {
				assert.True(t, resp.Ok, "Response ok should be true")
				assert.Equal(t, testKey, resp.Key, "ID key should be equal")
			}
		})
}

func TestGetJSONDoc(t *testing.T) {
	r := gofight.New()
	r.GET("/document/"+testKey).
		SetDebug(true).
		Run(engine, func(r gofight.HTTPResponse, rq gofight.HTTPRequest) {
			assert.Equal(t, http.StatusOK, r.Code)
			data := []byte(r.Body.String())
			var resp ResponseType
			err := json.Unmarshal(data, &resp)
			if assert.NoError(t, err) {
				assert.True(t, resp.Ok, "Response ok should be true")
				assert.JSONEq(t, testJSON, resp.Document)
				assert.Equal(t, testExtractor, resp.Extractor, "Extractor metadata should match")
				assert.Equal(t, testTitle, resp.Title, "Title metadata should match")
			}
		})
}

func TestDeleteJSONDoc(t *testing.T) {
	r := gofight.New()
	r.DELETE("/document/"+testKey).
		SetDebug(true).
		Run(engine, func(r gofight.HTTPResponse, rq gofight.HTTPRequest) {
			assert.Equal(t, http.StatusOK, r.Code)
			data := []byte(r.Body.String())
			var resp ResponseType
			err := json.Unmarshal(data, &resp)
			if assert.NoError(t, err) {
				assert.True(t, resp.Ok, "Response ok should be true")
			}
		})
}

func cleanupDoc(t *testing.T, key string) {
	errFile := os.Remove(dataDir + "/" + key)
	assert.NoError(t, errFile)
	errDB := db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(dbBucket).Delete([]byte(key))
	})
	assert.NoError(t, errDB)
}
