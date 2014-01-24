/*global module:true, require:true, console:true, process:true, async:true */

'use strict';

var path = require('path')
  , http = require('http')
  , queue = require('queue-async')
  , config = require('yaml-config')
  , Etcd = require('node-etcd');

var etcdHost = process.env.ETCD_HOST || 'localhost'
  , etcdPort = process.env.ETCD_PORT || '4001'
  , defaultConfig;

// if process.env.NODE_ENV has not been set, default to development
var NODE_ENV = process.env.NODE_ENV || 'development';


exports.load = load;


// override defaults if there is an etcd instance
function load(loadCallback) {
  // load defaults
  defaultConfig = config.readConfig(path.join(__dirname, 'defaultConfig.yml'));

  queue(1)
    .defer(checkEtcd)
    .defer(loadFromEtcd)
    .await(function (err, chk, etcdConfig) {
      var c = etcdConfig || defaultConfig;
      return loadCallback(null, c);
    });

}

function checkEtcd (callback) {
  http.get('http://' + etcdHost + ':' + etcdPort, function(res) {
    return callback(null);
  })
  .on('error', function (err) {
    console.error('etcd server, http://' + etcdHost + ':' + etcdPort + ', not available, using defaults');
    return callback(err);
  });
}

function loadFromEtcd (callback) {
  console.info('getting configuration from etcd');
  var etcd = new Etcd(etcdHost, etcdPort);

  queue(4)
    .defer(function (cb) {
      getValueFromEtcd(etcd, 'stucco/document-service/port', cb);
    })
    .defer(function (cb) {
      getValueFromEtcd(etcd, 'stucco/document-service/cluster', cb);
    })
    .defer(function (cb) {
      getListFromEtcd(etcd, 'riak/servers', cb);
    })
    .defer(function (cb) {
      getValueFromEtcd(etcd, 'stucco/document-service/riak/client', cb);
    })
    .defer(function (cb) {
      getValueFromEtcd(etcd, 'stucco/document-service/riak/pool', cb);
    })
    .defer(function (cb) {
      getValueFromEtcd(etcd, 'stucco/document-service/riak/bucket', cb);
    })
    .defer(function (cb) {
      getValueFromEtcd(etcd, 'stucco/document-service/logs/dir', cb);
    })
    .defer(function (cb) {
      getValueFromEtcd(etcd, 'stucco/document-service/logs/level', cb);
    })
    .await(function (err, port, cluster, rs, rc, rp, rb, logDir, logLevel) {
      if (err) {
        console.warn('error getting configuration from etcd, using defaults. Error: ');
        console.warn(err);
        return callback(err);
      }
      
      var cfg = defaultConfig;
      
      cfg.port = port;
      cfg.cluster = cluster;
      cfg.riak.servers = rs;
      cfg.riak.client = rc;
      cfg.riak.pool = rp;
      cfg.riak.bucket = rb;
      cfg.logs.dir = logDir;
      cfg.logs.level = logLevel;

      return callback(null, cfg);
    });

}

function getListFromEtcd (etcd, key, callback) {
  etcd.get(key, function(err, val) {
    if (err) return callback(err);
    if (val && val.node && val.node.nodes) {
      var nodes = val.node.nodes;
      var servers = [];
      for (var i = 0; i < nodes.length; i++) {
        servers.push(nodes[i].value);
      }
      return callback(null, servers);
    }
    else {
      return callback('no values');
    }
  });
}

function getValueFromEtcd (etcd, key, callback) {
  etcd.get(key, function(err, val) {
    if (err) return callback(err);
    if (val && val.node && val.node.value) {
      return callback(null, val.node.value);
    }
    else {
      return callback('no value');
    }
  });
}
