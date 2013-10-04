#!/bin/sh

# Specify the content-type in the HTTP header when adding documents,
#  that specifies the content-type in Riak

# Add with ID test
# Expected response: 
#  {"key":"23456","message":"document added"}%
curl -XPUT -i http://localhost:8000/add/23456 -H 'accept: application/json' -H 'content-type: application/json' -d '{ "a": 2, "b": 3 }'

# Add without ID test, a default ID will be returned
# Expected response (the key will actually be different): 
#  {"key":"JPNH05dUYjA8EH1eOBBdgTYei0g","message":"document added"}%
curl -XPUT -i http://localhost:8000/add -H 'accept: application/json' -H 'content-type: application/json' -d '{ "c": 4, "d": 5 }'

# Get test
# Expected response: 
#  {"a":2,"b":3}
curl -XGET -i http://localhost:8000/get/23456 -H 'accept: application/json' -H 'content-type: application/json'

