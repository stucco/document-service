document-service
================

*API for interacting with raw document store*

## Configuration

See `config.yaml` for setting the server's listen port, log information, and [riak](http://basho.com/riak/) configuration.


## Usage

```bash
# Specify the content-type in the HTTP header when adding documents,
#  that specifies the content-type in Riak

# Add with ID test
# Expected response: 
#  {"key":"23456","message":"document added"}%
curl -XPUT -i http://localhost:8000/document/23456 -H 'accept: application/json' -H 'content-type: application/json' -d '{ "a": 2, "b": 3 }'

# Add without ID test, a default ID will be returned
# Expected response (the key will actually be different): 
#  {"key":"JPNH05dUYjA8EH1eOBBdgTYei0g","message":"document added"}%
curl -XPUT -i http://localhost:8000/document -H 'accept: application/json' -H 'content-type: application/json' -d '{ "c": 4, "d": 5 }'

# Get test
# Expected response: 
#  {"a":2,"b":3}
curl -XGET -i http://localhost:8000/document/23456 -H 'accept: application/json' -H 'content-type: application/json'
```

## Install

Just install the required modules.

    npm install -d

## Running

To start:

    npm start

To set the `etcd` host and port, run:

    ETCD_HOST=127.0.0.1 ETCD_PORT=4001 npm start

If `etcd` is unavailable, it will use the settings defined in the config file.


## Logging

Logs will be put in the `logs` directory as JSON files. Control the log-level by setting `logs.level` in `config.yaml`. Errors will also be put in a separate file.


## Testing

```bash
npm test
```


## Docker and Vagrant

To build a virtual machine with [Vagrant](http://www.vagrantup.com/) to expose the `document-service` as a Linux container using [Docker](http://www.docker.io/), do the following to get the VM up and running:

```bash
vagrant up
vagrant ssh
```

Once logged into the VM, build the docker image and start the container (the ports in the run command should match the port set in the `config.yaml` and in the `Dockerfile`):

```bash
sudo docker build -t stucco/document-service .
sudo docker run -p 8000:8000 stucco/document-service%        
```


## License

This software is freely distributable under the terms of the MIT License.

Copyright (c) UT-Battelle, LLC (the "Original Author")

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:
 
The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
 
THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS, THE U.S. GOVERNMENT, OR UT-BATTELLE BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
