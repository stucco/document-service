#!/bin/sh

# Get test
curl -XGET -i http://localhost:8000/get/23456 -H 'accept: application/json' -H 'content-type: application/json'

# Add test
curl -XPUT -i http://localhost:8000/add -H 'accept: application/json' -H 'content-type: application/json' -d '
{
  "a": 2,
  "b": 3
}
'