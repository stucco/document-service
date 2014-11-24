document-service
================

[![Gobuild Download](http://gobuild.io/badge/github.com/stucco/document-service/downloads.svg)](http://gobuild.io/github.com/stucco/document-service)

Store text files and metadata with an HTTP API

## API

The API is exposed on `host:port/document/` with the following routes:

* *Get a document*: `GET host:port/document/<id>`. It returns a JSON object that describes the success or failure.
* *Post a document*: `POST host:port/document/` will assign an id, or `POST host:port/document/<id>` to specify the id. It returns a JSON object that describes the success or failure.
* *Delete a document*: `DELETE host:port/document/<id>`. It returns a JSON object that describes the success or failure.

#### Examples

Below are examples using [`curl`](http://curl.haxx.se).

Upload  a json file:

```
curl -XPOST localhost:8000/document/12345\?extractor=test --data "{key1: 'some data', key2: 'more data'}" -i -H "Content-Type: application/json"
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