/*global describe:true, before:true, after: true, it:true */

'use strict';

var util = require('util')
  , path = require('path')
  , fs = require('fs')
  , queue = require('queue-async')
  , should = require('chai').should()
  , request = require('supertest')
  , riak = require('riak-js');

var settings = require('yaml-config').readConfig(path.join(__dirname, '..', 'defaultConfig.yml'), 'default')
  , url = 'http://127.0.0.1:' + settings.port
  , riakBucket
  , riakClient;

// testing parameters for 'get' tests
var getTestOneKey = 'testJson12'
  , getTestOneData = {v1: 30, v2: 45}
  , getTestTwoKey = 'testJson34'
  , getTestTwoData = {v1: 40, v2: 55}
  , getTestTwoBucket = 'testJson56'
  , getTestThreeKey = 'testPdf12'
  , getTestThreeFile = path.join(__dirname, 'fixtures', 'test.pdf')
  , getTestThreeData = fs.readFileSync(getTestThreeFile, {encoding: null})
  , putTestBucket = 'testBucket00';


describe('Test routes (' + url + ')', function () {

  before(function (done) {
    
    // add document to riak
    var servers = settings.riak.servers || ['localhost:8098']
      , client = 'riak-js-test'
      , pool = 'test-pool';

    riakBucket = settings.riak.bucket || 'testBucket';

    // open riak connection
    riakClient = riak.getClient({pool: {servers: servers, name: pool, keepAlive: true, encodeUri: true}, clientId: client});

    queue(3)
      .defer(function(callback) {
        riakClient.save(riakBucket, getTestOneKey, getTestOneData, {} , function(err) {
          return callback(err);
        });
      })
      .defer(function(callback) {
        riakClient.save(getTestTwoBucket, getTestTwoKey, getTestTwoData, {} , function(err) {
          return callback(err);
        });
      })
      .defer(function(callback) {
        riakClient.save(riakBucket, getTestThreeKey, getTestThreeData, {contentType: 'application/pdf'} , function(err) {
          return callback(err);
        });
      })
      .awaitAll(function (err) {
        if (err) return done(err);
        request(url)
            .get('/')
            .expect(404)
            .end(function (err, res) {
              if (err) {
                if (err.code === 'ECONNREFUSED') return done(new Error('Server is not running.'));
                return done(err);
              }
              return done();
            });
      });

  });

  describe('Test routes: errors', function () {

    it('should return not found error (route:  /noroute)', function (done) {
      request(url)
            .get('/noroute')
            .set('accept', 'application/json')
            .expect(404)
            .end(function (err, res) {
              return done();
            });
    });

    it('should return key not found error (route: GET /document/:key)', function (done) {
      request(url)
            .get('/document/asdfkhag87y')
            .set('accept', 'application/json')
            .expect(404)
            .end(function (err, res) {
              var response = res.body;
              response.code.should.equal('ResourceNotFound');
              return done();
            });
    });

  });

  describe('Test routes: get JSON documents', function () {

    it('should return the correct JSON document (route: GET /document/:key)', function (done) {
      request(url)
            .get('/document/' + getTestOneKey)
            .set('accept', 'application/json')
            .expect(200)
            .end(function (err, res) {
              if (err) return done(err);
              var response = JSON.parse(res.text);
              response.should.be.an('object');
              response.v1.should.equal(getTestOneData.v1);
              response.v2.should.equal(getTestOneData.v2);
              return done();
            });
    });

    it('should return the correct JSON document from a specified bucket (route: GET /document/:key?bucket=bucketName)', function (done) {
      request(url)
            .get('/document/' + getTestTwoKey + '?bucket=' + getTestTwoBucket)
            .set('accept', 'application/json')
            .expect(200)
            .end(function (err, res) {
              if (err) return done(err);
              var response = JSON.parse(res.text);
              response.should.be.an('object');
              response.v1.should.equal(getTestTwoData.v1);
              response.v2.should.equal(getTestTwoData.v2);
              return done();
            });
    });

  });

  describe('Test routes: add JSON documents', function () {

    it('should add a JSON document and return a key (route: PUT /document)', function (done) {
      var doc = {'testkey1': 'testval'};
      request(url)
            .put('/document')
            .set('accept', 'application/json')
            .send(doc)
            .expect(200)
            .end(function (err, res) {
              if (err) return done(err);
              var response = res.body
                , generatedKey = response.key;
              response.should.be.an('object');
              should.exist(generatedKey);
              response.message.should.equal('document added');
              riakClient.get(riakBucket, generatedKey, {}, function(riakErr, data) {
                if (riakErr) return done(riakErr);
                data.should.deep.equal(doc);
                return done();
              });
            });
    });

    it('should add a JSON document into a specified bucket and return a key (route: PUT /document?bucket=bucketName)', function (done) {
      var doc = {'testkey2': 'testval'};
      request(url)
            .put('/document?bucket=' + putTestBucket)
            .set('accept', 'application/json')
            .send(doc)
            .expect(200)
            .end(function (err, res) {
              if (err) return done(err);
              var response = res.body
                , generatedKey = response.key;
              response.should.be.an('object');
              should.exist(generatedKey);
              response.message.should.equal('document added');
              riakClient.get(putTestBucket, generatedKey, {}, function(riakErr, data) {
                if (riakErr) return done(riakErr);
                data.should.deep.equal(doc);
                return done();
              });
            });
    });

    it('should add a JSON document with a defined key (route: PUT /document/:key)', function (done) {
      var key = 'test87654'
        , doc = {'testkey3': 'testval'};
      request(url)
            .put('/document/' + key)
            .set('accept', 'application/json')
            .send(doc)
            .expect(200)
            .end(function (err, res) {
              if (err) return done(err);
              var response = res.body
                , generatedKey = response.key;
              response.should.be.an('object');
              should.exist(generatedKey);
              generatedKey.should.equal(key);
              response.message.should.equal('document added');
              riakClient.get(riakBucket, key, {}, function(riakErr, data) {
                if (riakErr) return done(riakErr);
                data.should.deep.equal(doc);
                return done();
              });
            });
    });

    it('should add a JSON document into a specified bucket with a defined key (route: PUT /document/:key?bucket=bucketName)', function (done) {
      var doc = {'testkey4': 'testval'}
        , key = 'test87654';
      request(url)
            .put('/document/' + key + '?bucket=' + putTestBucket)
            .set('accept', 'application/json')
            .send(doc)
            .expect(200)
            .end(function (err, res) {
              if (err) return done(err);
              var response = res.body
                , generatedKey = response.key;
              response.should.be.an('object');
              should.exist(generatedKey);
              response.message.should.equal('document added');
              riakClient.get(putTestBucket, generatedKey, {}, function(riakErr, data) {
                if (riakErr) return done(riakErr);
                data.should.deep.equal(doc);
                return done();
              });
            });
    });

  });

  describe('Test routes: get PDF documents', function () {

    // restify ships with formatters for application/json, text/plain,
    // and application/octet-stream. do not try to use a more specific
    // accept format (e.g. application/pdf)
    // @see http://mcavage.me/node-restify/#Content-Negotiation
    it('should return the correct PDF document (route: GET /document/:key)', function (done) {
      request(url)
            .get('/document/' + getTestThreeKey)
            .set('accept', 'application/octet-stream')
            .expect(200)
            .end(function (err, res) {
              if (err) return done(err);
              var data = res.text;

              //TODO - TEST IF THE RETURNED PDF IS THE SAME
              //data.should.deep.equal(getTestThreeData);

              return done();
            });
    });

  });

  describe('Test compression', function () {

    it('should return the correct JSON document (route: GET /document/:key)', function (done) {
      request(url)
            .get('/document/' + getTestOneKey)
            .set('accept', 'application/json')
            .set('accept-encoding', 'application/gzip')
            .expect(200)
            .end(function (err, res) {
              if (err) return done(err);
              var response = JSON.parse(res.text);
              response.should.be.an('object');
              response.v1.should.equal(getTestOneData.v1);
              response.v2.should.equal(getTestOneData.v2);
              return done();
            });
    });

  });

});