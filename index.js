/* globals require, console, process */
'use strict';

var util = require('util');

var RiakClient = require("riak");

// list of riak servers you'd like to load balance over (poolee handles this).
var riakServers = ["localhost:8098"];

// should be unique, used by riak if you don't supply a vector clock
// default value: random integer
var riakClientId = "stucco-document-service-client";

// informative name for logging purposes etc
var riakPoolName = "stucco-document-service-pool";

// riak bucket
var riakBucket = "stucco";

var riakClient = new RiakClient(riakServers, riakClientId, riakPoolName);

// shows an activity trace.
//riakClient.debug_mode = true;


riakClient.put(riakBucket, 'testkey', {testvalue1: 30, testvalue2: 45}, {} , function(error, response, result) {
  if (error) console.error(error);

  // console.log('response: ' + response);
  console.log('put result: ' + util.inspect(result));
});


riakClient.get(riakBucket, 'testkey', {}, function(error, response, result) {
  if (error) console.error(error);

  // console.log('response: ' + util.inspect(response));
  console.log('get result: ' + util.inspect(result));
  process.exit(0);
});