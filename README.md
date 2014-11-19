document-service
================

## Usage

## API

The API is exposed on `host:port/document/` with the following routes:

* *Get a document*: `GET host:port/document/<id>`. It returns the requested document in the response body if the document exists, or a JSON object if there is an error.
* *Post a document*: `POST host:port/document/` will assign an id, or `POST host:port/document/<id>` to specify the id. It returns a JSON object that describes the success or failure.
* *Delete a document*: `DELETE host:port/document/<id>`. It returns a JSON object that describes the success or failure.

#### Examples

Below are examples using [`curl`](http://curl.haxx.se). To see examples of a [golang](http://golang.org/) application, see the example in the `client` directory.

Upload  a json file:

```
curl -XPOST localhost:8000/document/12345 --data "{key1: 'some data', key2: 'more data'}" -i -H "Content-Type: application/json"
HTTP/1.1 200 OK
Content-Type: application/json
Date: Wed, 19 Nov 2014 00:06:45 GMT
Content-Length: 63

{"key":"12345","message":"document added with id","ok":"true"}
```

Upload a binary file:
```
curl -XPOST localhost:8000/document/ --data-binary "@/Users/ojg/Downloads/DetailPage.png" -i -H "Content-Type: application/octet-stream"
HTTP/1.1 100 Continue

HTTP/1.1 200 OK
Content-Type: application/json
Date: Wed, 19 Nov 2014 00:07:42 GMT
Content-Length: 86

{"key":"68b24d2d-458b-42d9-9c55-b76dc4737ca3","message":"document added","ok":"true"}
```

Retrieve a file:
```
    curl -XGET -O localhost:8000/document/68b24d2d-458b-42d9-9c55-b76dc4737ca3 -s
```