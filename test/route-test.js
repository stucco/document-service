/*global describe:true, before:true, after: true, it:true */

'use strict';

var util = require('util')
  , path = require('path')
  , fs = require('fs')
  , settings = require('yaml-config').readConfig(path.join(__dirname, '..', 'config.yaml'), 'default')
  , should = require('chai').should()
  , request = require('supertest')
  , url = 'http://localhost:' + settings.server.port
  , defaultBucket
  , riak = require('riak-js')
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


describe('Test routes', function () {

  before(function (done) {
    
    var server = require('../server');

    // add document to riak
    var servers = settings.riak.servers || ['localhost:8098']
      , client = 'riak-js-test'
      , pool = 'test-pool';

    defaultBucket = settings.riak.bucket || 'testBucket';

    // open riak connection
    riakClient = riak.getClient({pool: {servers: servers, name: pool, keepAlive: true, encodeUri: true}, clientId: client});

    // prepare simple get JSON test
    riakClient.save(defaultBucket, getTestOneKey, getTestOneData, {} , function(error) {
      if (error) console.error(error);
    });

    // prepare custom bucket JSON test
    riakClient.save(getTestTwoBucket, getTestTwoKey, getTestTwoData, {} , function(error) {
      if (error) console.error(error);
    });

    // prepare PDF test
    riakClient.save(defaultBucket, getTestThreeKey, getTestThreeData, {contentType: 'application/pdf'} , function(error) {
      if (error) console.error(error);
    });


    // make sure the server is started
    setTimeout(function() {
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
    }, 1500);
  });

  describe('Test routes: errors', function () {

    it('should return not found error (route:  /noroute)', function (done) {
      request(url)
            .get('/get/noroute')
            .set('accept', 'application/json')
            .expect(404)
            .end(function (err, res) {
              var response = res.body;
              response.code.should.equal('ResourceNotFound');
              return done();
            });
    });

    it('should return key not found error (route:  /get/:key)', function (done) {
      request(url)
            .get('/get/asdfkhag87y')
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

    it('should return the correct JSON document (route:  /get/:key)', function (done) {
      request(url)
            .get('/get/' + getTestOneKey)
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

    it('should return the correct JSON document from a specified bucket (route:  /get/:key?bucket=bucketName)', function (done) {
      request(url)
            .get('/get/' + getTestTwoKey + '?bucket=' + getTestTwoBucket)
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

    it('should add a JSON document and return a key (route:  /add)', function (done) {
      var doc = {'testkey1': 'testval'};
      request(url)
            .put('/add')
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
              riakClient.get(defaultBucket, generatedKey, {}, function(riakErr, data) {
                if (riakErr) return done(riakErr);
                data.should.deep.equal(doc);
                return done();
              });
            });
    });

    it('should add a JSON document into a specified bucket and return a key (route:  /add?bucket=bucketName)', function (done) {
      var doc = {'testkey2': 'testval'};
      request(url)
            .put('/add?bucket=' + putTestBucket)
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

    it('should add a JSON document with a defined key (route:  /add/:key)', function (done) {
      var key = 'test87654'
        , doc = {'testkey3': 'testval'};
      request(url)
            .put('/add/' + key)
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
              riakClient.get(defaultBucket, key, {}, function(riakErr, data) {
                if (riakErr) return done(riakErr);
                data.should.deep.equal(doc);
                return done();
              });
            });
    });

    it('should add a JSON document into a specified bucket with a defined key (route:  /add/:key?bucket=bucketName)', function (done) {
      var doc = {'testkey4': 'testval'}
        , key = 'test87654';
      request(url)
            .put('/add/' + key + '?bucket=' + putTestBucket)
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
    it('should return the correct PDF document (route:  /get/:key)', function (done) {
      request(url)
            .get('/get/' + getTestThreeKey)
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

    it('should return the correct JSON document (route:  /get/:key)', function (done) {
      request(url)
            .get('/get/' + getTestOneKey)
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