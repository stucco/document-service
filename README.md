document-service
================

*API for interacting with raw document store*

## Configuration

See `config.yaml` for setting the server's listen port, log information, and [riak](http://basho.com/riak/) configuration.


## Usage

The script can spawn as many processes as there are cores using the nodejs [cluster module](http://nodejs.org/docs/latest/api/cluster.html). To run only a single worker, set `server.cluster: false` in `config.yaml`.

    # start the cluster
    npm start
    # stop the cluster
    npm stop

`package.json` is set up to use [forever](https://github.com/nodejitsu/forever) to manage processes, so `npm start` and `npm stop` will run forever to start the `server.js` script. `npm run-script list` will list out the forever processes that are running.


## Logging

Logs will be put in the `logs` directory as JSON files. Control the log-level by setting `logs.level` in `config.yaml`. Errors will also be put in a separate file.


## Testing

    npm test


## License

This software is freely distributable under the terms of the MIT License.

Copyright (c) UT-Battelle, LLC (the "Original Author")

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:
 
The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
 
THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS, THE U.S. GOVERNMENT, OR UT-BATTELLE BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
