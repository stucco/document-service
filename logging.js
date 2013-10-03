/*global module:true, require:true */

'use strict';

var fs = require('fs')
  , path = require('path')
  , config = require('yaml-config')
  , bunyan = require('bunyan');


exports.createLogger = createLogger;


/*
 * configure and start logging
 * @param {Object} config The configuration object for defining dir: log directory, level: loglevel
 * @return the created logger instance
 */
function createLogger (config) {

  var pkg = require(path.join(__dirname, 'package'))
    , appName = pkg.name
    , appVersion = pkg.version
    , logDir = config.dir || path.join(__dirname, 'logs')
    , logFile = path.join(logDir, appName + '-log.json')
    , logErrorFile = path.join(logDir, appName + '-errors.json')
    , logLevel = config.level || 'debug';

  // Create log directory if it doesnt exist
  if (! fs.existsSync(logDir)) fs.mkdirSync(logDir);

  // Log to console and log file
  var log = bunyan.createLogger({
    name: appName
  , streams: [ 
      {
        stream: process.stdout
      , level: 'warn'
      }
    , { 
        path: logFile
      , level: logLevel
      , type: 'rotating-file'
      , period: '1d'
      }
    , { 
        path: logErrorFile
      , level: 'error'
      }
    ]
  , serializers: bunyan.stdSerializers
  });

  log.info('Starting ' + appName + ', version ' + appVersion);
  log.info('Environment set to ' + process.env.NODE_ENV);
  log.debug('Logging setup completed.');
  
  return log;
}
