/*global module:true, require:true, console:true, process:true */

'use strict';

var path = require('path')
  , restify = require('restify')
  , riak = require('riak-js')
  , riakClient;

exports.createServer = createServer;

/*
 * Set up server
 * @return the created server
 */
function createServer (logger, riakConfig) {

  var config = {
    name: require(path.join(__dirname, 'package')).name
  };

  if (logger) config.log = logger;

  // create riak client connection (uses poolee for pooling)
  riakClient = riak.getClient({pool: {servers: riakConfig.servers, name: riakConfig.pool, keepAlive: true, encodeUri: true}, clientId: riakConfig.client});

  // create restify server
  var server = restify.createServer(config);
  server.use(restify.acceptParser(server.acceptable));
  server.use(restify.queryParser());
  server.use(restify.gzipResponse());

  // default not found route
  server.on('NotFound', function (req, res, next) {
    if (logger) logger.debug('404', 'Request for ' + req.url + ' not found. No route.');
    res.send(404, req.url + ' was not found');
  });
  
  if (logger) server.on('after', restify.auditLogger({ log: logger }));
  
  // ROUTES
  
  // GET /document/:key -- retrieve a document based on an key
  // optionally add a bucket query parameter
  // Example: /document/2345
  server.get('/document/:key', _getDocument);

  // PUT /document -- add a document and return an key
  // optionally add a bucket query parameter
  // Example: /document or /document?bucket=bucketName
  server.put('/document', restify.bodyParser({mapParams: false}), _addDocument);

  // PUT /document/:key -- add a document using a given key
  // optionally add a bucket query parameter
  // Example: /document/23456 or /document/23456?bucket=bucketName
  server.put('/document/:key', restify.bodyParser({mapParams: false}), _addDocument);
  
  
  // retrieve document from riak
  // key parameter is required
  function _getDocument (req, res, next) {
    if (! req.params.key) {
      return next(new restify.MissingParameterError('Supply a document key'));
    }
    var key = req.params.key
      , bucket = req.params.bucket || riakConfig.bucket;

    // set the content type to json
    res.contentType = 'json';

    // retrieve the document
    riakClient.get(bucket, key, {stream: true}, function(error, data, meta) {
      if (error) {
        if (error.statusCode === 404) {
          logger.warn('Riak key \'' + key + '\' not found.');
          return next(new restify.ResourceNotFoundError('key ' + key + ' not found.'));
        }
        else {
          logger.error(error);
          return next(new restify.InternalError('Riak GET error: ' + error));
        }
      }

      // pass on the content-type from riak to the response
      res.contentType = meta.contentType;

      // stream to response (streamed to response.text)
      data.pipe(res)
      .on('error', function(err) {
        logger.error(error);
        return next(new restify.InternalError('Riak GET error: ' + error));
      })      
      .on('end', function() {
        logger.debug('Retrieved document \'' + key + '\' from Riak');
        return next();
      });

    });
  }


  // add document to riak
  // TODO - USE STREAMS
  function _addDocument (req, res, next) {

    var doc = req.body
      , bucket = req.params.bucket || riakConfig.bucket
      , type = req.contentType
      , length = req.contentLength;

    if (! doc) {
      return next(new restify.MissingParameterError('Supply the document to add'));
    }

    if (type === 'application/json' && typeof doc === 'object') {
      doc = JSON.stringify(doc);
    }

    // key is either defined in the request, or will be assigned by riak
    riakClient.save(bucket, req.params.key, doc, function(error, response, meta) {
      var key = meta.key;
      if (error) {
        logger.error(error);
        return next(new restify.InternalError('Riak PUT error: ' + error));
      }

      logger.debug('Added document to Riak with key \'' + key + '\'');

      res.send({'key': key, 'message': 'document added'});
      return next();
    });

  }

  return server;
}
