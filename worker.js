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
  riakClient = riak.getClient({pool: {servers: riakConfig.servers, name: riakConfig.pool, keepAlive: true}, clientId: riakConfig.client});

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
  
  // retrieve a document based on an key
  // /get/:key
  // USAGE EXAMPLE: /get/2345
  server.get('/get/:key', _getDocument);

  // add a document and return an key
  // /get/:key
  // USAGE EXAMPLE: /get/2345
  server.put('/add', restify.bodyParser({mapParams: false}), _addDocument);
  
  
  // retrieve document from riak
  // key parameter is required
  function _getDocument (req, res, next) {
    if (! req.params.key) {
      return next(new restify.MissingParameterError('Supply a document key'));
    }
    var key = req.params.key;

    // set the content type to json
    res.contentType = 'json';

    // retrieve the document
    riakClient.get(riakConfig.bucket, key, {stream: true}, function(error, data, meta) {
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

      // stream to response (streamed to response.text)
      data.pipe(res)
      .on('error', function(err) {
        logger.error(error);
        return next(new restify.InternalError('Riak GET error: ' + error));
      })      
      .on('end', function() {
        return next();
      });

    });
  }


  // add document to riak
  function _addDocument (req, res, next) {

    var doc = req.body
      , type = req.contentType
      , length = req.contentLength;
    if (type === 'application/json' && typeof doc === 'object') {
      doc = JSON.stringify(doc);
    }

    riakClient.save(riakConfig.bucket, null, doc, function(error, response, meta) {
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
