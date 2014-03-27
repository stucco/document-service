/*global module:true, require:true, console:true, process:true */

'use strict';

var path = require('path')
  , fs = require('fs')
  , restify = require('restify')
  , uuid = require('uuid')
  , mime = require('mime')
  , mkdirp = require('mkdirp')
  , tika = require('tika');

exports.createServer = createServer;

/*
 * Set up server
 * @return the created server
 */
function createServer (logger, config) {

  var restifyOpts = {
    name: require(path.join(__dirname, 'package')).name
  };

  if (logger) restifyOpts.log = logger;

  // check that data directory exists
  mkdirp.sync(config.data.dir, '0777');

  // create restify server
  var server = restify.createServer(restifyOpts);
  server.use(restify.acceptParser(server.acceptable));
  server.use(restify.queryParser());
  server.use(restify.gzipResponse());

  // default not found route
  server.on('NotFound', function (req, res, next) {
    if (logger) logger.debug('404', 'Request for ' + req.url + ' not found. No route.');
    res.send(404, 'Route \'' + req.url + '\' was not found');
  });
  
  if (logger) server.on('after', restify.auditLogger({ log: logger }));
  
  // ROUTES
  
  // GET /document/:key -- retrieve a document based on an key
  // Example: /document/2345
  server.get('/document/:key', _getDocument);

  // PUT /document -- add a document and return an key
  // Example: /document or /document?bucket=bucketName
  server.put('/document', _putDocument);

  // PUT /document/:key -- add a document using a given key
  // Example: /document/23456 or /document/23456
  server.put('/document/:key', _putDocument);
  
  
  // GET document
  // key parameter is required
  function _getDocument (req, res, next) {
    if (! req.params.key) {
      return next(new restify.MissingParameterError('Supply a document key'));
    }

    var key = req.params.key
      , filePath = path.join(config.data.dir, key);

    if (req.params.extract) {
      tika.extract(filePath, function(err, text, meta) {
        if (err) {
          logger.error(err);
          return next(new restify.InternalError(err));
        }
        res.send(200, {'text': text, 'meta': meta});
        return next();
      });
    }
    else {
      res.contentType = mime.lookup(filePath);

      // get the file stream to read from
      var inStream = fs.createReadStream(filePath); //, {encoding: 'utf8'});
      inStream.on('error', function(err) {
        logger.error(err);
        if (err.code === 'ENOENT') {
          return next(new restify.ResourceNotFoundError('key ' + key + ' not found.'));
        }
        else {
          return next(new restify.InternalError(err));
        }
      })
      .on('end', function() {
        return next();
      });

      // send content
      inStream.pipe(res);
    }

  }


  // PUT document
  function _putDocument (req, res, next) {

    // get a key or use the one defined
    var key = req.params.key ? req.params.key : uuid.v4()
      , filePath = path.join(config.data.dir, key);

    var outStream = fs.createWriteStream(filePath); //, {encoding: 'utf8'});

    req.on('error', function(err) {
      logger.error(err);
      return next(new restify.InternalError(err));
    })
    .on('end', function() {
      logger.debug('Added document with key \'' + key + '\'');
      res.send(200, {'key': key, 'message': 'document added'});
      return next();
    });

    req.pipe(outStream);
  }

  return server;
}
