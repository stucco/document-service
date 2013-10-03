restify-base
============

## Boilerplate setup for building a [Node.js](http://nodejs.org/) RESTful API server with [restify](http://mcavage.github.io/node-restify/).


## Included

This project will set up the basic structure for restify including setting up a [YAML configuration file](https://github.com/rjyo/yaml-config-node), logging with [bunyan](https://github.com/trentm/node-bunyan), and testing with [mocha](http://visionmedia.github.io/mocha/) and [supertest](https://github.com/visionmedia/supertest). The server will spawn a worker for each CPU using node's [cluster module](http://nodejs.org/docs/latest/api/cluster.html).


## Usage

`package.json` is set up to use [forever](https://github.com/nodejitsu/forever) to manage processes, so `npm start` and `npm stop` will run forever to start the `server.js` script. `npm run-script list` will list out the forever processes that are running.

1. Customize the [`package.json`](https://npmjs.org/doc/files/package.json.html) file with appropriate name, version, and dependency versions.
2. Install dependencies: `npm install -d`
3. Run it: `npm start`

## Testing

There is an example of testing the example route in the `test` directory. Run all test specifications with mocha using: `npm test`
