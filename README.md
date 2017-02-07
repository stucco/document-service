document-service
================

[![Build Status](https://travis-ci.org/stucco/document-service.svg?branch=master)](https://travis-ci.org/stucco/document-service)

This software provides a storage service for text documents and metadata over an HTTP API.

## Usage

Start the service (use `-h` for options):

    ./doc-service

## API

The API is exposed on `host:port/document/` with the following routes:

* *Get a document*: `GET host:port/document/<id>`. It returns a JSON object of the document and meta-data and that also describes the success or failure.
* *Post a document*: `POST host:port/document/` will assign an id, or `POST host:port/document/<id>` to specify the id. It returns a JSON object that describes the success or failure.
* *Delete a document*: `DELETE host:port/document/<id>`. It returns a JSON object that describes the success or failure.

### Examples

Below are examples using [`curl`](http://curl.haxx.se).

#### JSON Data

Upload a json file:

```
curl -XPOST localhost:8000/document/12345\?extractor\=test\&title\=test --data "{key1: 'some data', key2: 'more data'}" -i -H "Content-Type: application/json"
HTTP/1.1 200 OK
Content-Type: application/json
Date: Fri, 21 Nov 2014 01:53:18 GMT
Content-Length: 61

{"ok":"true","key":"12345","message":"saved document by id"}
```

Retrieve a file:
```
curl -XGET localhost:8000/document/12345 -i
HTTP/1.1 200 OK
Content-Type: application/json
Date: Fri, 21 Nov 2014 01:54:41 GMT
Content-Length: 122

{"ok":"true","key":"12345","document":"{key1: 'some data', key2: 'more data'}","timestamp":1416534798,"extractor":"test"}
```

Delete a file:
```
curl -XDELETE localhost:8000/document/12345 -i
HTTP/1.1 200 OK
Content-Type: application/json
Date: Fri, 21 Nov 2014 01:55:38 GMT
Content-Length: 57

{"ok":"true","key":"12345","message":"removed document"}
```

#### Binary Data

Upload an image file:
```
base64 file.png | curl -XPOST localhost:8000/document/ --data @- -i -H "Content-Type: image/png"
HTTP/1.1 100 Continue

HTTP/1.1 200 OK
Content-Type: application/json
Date: Thu, 15 Jan 2015 20:36:16 GMT
Content-Length: 86

{"ok":"true","key":"befc3e40-e3de-4666-b7b5-155e1b0935d6","message":"saved document"}
```

Download an image file; this example uses [jq](http://stedolan.github.io/jq/) to extract the base64 data from the JSON object:
```
curl -XGET localhost:8000/document/1de60b72-e91b-4a26-9466-86f0d3ccdf7f --silent | jq --raw-output .document | base64 -D > file.png
```


## Development

To build for the current OS/arch:

```
go build doc-service.go
```

To build for mac, linux and windows on 64 bit:

```
oses=( darwin linux windows )
for os in "${oses[@]}"; do
  out="doc-service_${os}_amd64"
  echo "Building ${out}..."
  env CGO_ENABLED=0 GOOS=$os GOARCH=amd64 go build -o $out -a -tags netgo
done
```

To remove all existing builds:

```
rm doc-service_{darwin,linux,windows}_amd64
```

To test:

```
go test .
```
