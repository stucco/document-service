/*global module:true, require:true, console:true, process:true */

'use strict';

var path = require('path')
  , restify = require('restify')
  , Riak = require('riak');


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

  // create riak client connection
  var riakClient = new Riak(riakConfig.servers, riakConfig.client, riakConfig.pool);

  // create restify server
  var server = restify.createServer(config);

  server.use(restify.acceptParser(server.acceptable));
  server.use(restify.queryParser());
  server.use(restify.gzipResponse());


  server.on('NotFound', function (req, res, next) {
    if (logger) logger.debug('404', 'Request for ' + req.url + ' not found. No route.');
    res.send(404, req.url + ' was not found');
  });
  
  if (logger) server.on('after', restify.auditLogger({ log: logger }));
  
  // ROUTES
  
  // retrieve a document based on an id
  // /get/:id
  // USAGE EXAMPLE: /get/2345
  server.get('/get/:id', _getDocument);

  // add a document and return an id
  // /get/:id
  // USAGE EXAMPLE: /get/2345
  server.put('/add', _addDocument);
  
  
  function _getDocument (req, res, next) {
    var id = req.params.id;

    // retrieve document from riak
    riakClient.get(riakConfig.bucket, id, {}, function(error, response, result) {
      if (error) {
        logger.error(error);
        return next(new restify.InternalError('Riak error: ' + error));
      }

      if (!result) {
        var riakErrMsg = 'Riak error getting document ' + id;
        logger.error(riakErrMsg);
        return next(new restify.InternalError(riakErrMsg));
      }
      else {
        res.send(result);
        return next();
      }
    });
  }

  function _addDocument (req, res, next) {

    var fileName = req.params.name;
    var fileBytes = req.headers['content-length'];
    var newFile = require('fs').createWriteStream('./uploads/' + fileName);
    var uploadedBytes = 0;
   
    console.log(fileBytes);
    console.log(req.headers);
   
    req.on('end', function(){
      console.log("on finish event");
      res.send(200, { message: "File uploaded" });
      return next();
    });
   
    req.on('data', function(chunk){
      console.log("On the data event");
      uploadedBytes += chunk.length;
      var progress = (uploadedBytes / fileBytes) * 100;
      console.log(progress);
      console.log(uploadedBytes);
    });
   
    req.pipe(newFile);

  }

  return server;
}
