/*global describe:true, before:true, after: true, it:true */

'use strict';

var util = require('util')
  , path = require('path')
  , fs = require('fs')
  , should = require('chai').should()
  , request = require('supertest');

var settings = require('yaml-config').readConfig(path.join(__dirname, '..', 'defaultConfig.yml')) // use NODE_ENV to specify environment to use, else default
  , url = 'http://127.0.0.1:' + settings.port;

var filesToCleanup = [];

// testing parameters for 'get' tests
var getTestOneKey = 'testJson12'
  , getTestOneData = {v1: 30, v2: 45}
  , getTestTwoKey = 'testXml34'
  , getTestTwoData = '<xml><test>a test</test></xml>'
  , getTestThreeKey = 'testPdf12'
  , getTestThreeFile = path.join(__dirname, 'fixtures', 'test.pdf')
  , getTestThreeData = fs.readFileSync(getTestThreeFile, {encoding: null});


describe('Test routes (' + url + ')', function () {

  before(function (done) {

    // set up files to get
    var getTestOnePath = path.join(settings.data.dir, getTestOneKey);
    fs.writeFileSync(getTestOnePath, JSON.stringify(getTestOneData)); //, {encoding: 'utf8'});
    filesToCleanup.push(getTestOneKey);

    var getTestTwoPath = path.join(settings.data.dir, getTestTwoKey);
    fs.writeFileSync(getTestTwoPath, getTestTwoData); //, {encoding: 'utf8'});
    filesToCleanup.push(getTestTwoKey);

    var getTestThreePath = path.join(settings.data.dir, getTestThreeKey);
    fs.writeFileSync(getTestThreePath, getTestThreeData); //, {encoding: 'utf8'});
    filesToCleanup.push(getTestThreeKey);

    // start the server if it is not running
    request(url)
        .get('/')
        .timeout(500)
        .expect(404)
        .end(function (err, res) {
          if (err && err.code === 'ECONNREFUSED') {
            require('../server');
          }
          return done();
        });

  });

  after(function(done) {
    for (var i = 0; i < filesToCleanup.length; i++) {
      fs.unlink(path.join(settings.data.dir, filesToCleanup[i]));
    }
    return done();
  });

  describe('Test routes: errors', function () {

    it('should return not found error (route:  /noroute)', function (done) {
      request(url)
            .get('/noroute')
            .set('content-type', 'application/json')
            .set('accept', 'application/json')
            .expect(404)
            .end(function (err, res) {
              return done();
            });
    });

    it('should return key not found error (route: GET /document/:key)', function (done) {
      request(url)
            .get('/document/asdfk73656da532hag87y')
            .set('accept', 'application/json')
            .expect(404)
            .end(function (err, res) {
              return done();
            });
    });

  });


  describe('Test routes: get documents', function () {

    it('should return the correct JSON document (route: GET /document/:key)', function (done) {
      request(url)
            .get('/document/' + getTestOneKey)
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

    it('should return the correct XML document (route: GET /document/:key)', function (done) {
      request(url)
            .get('/document/' + getTestTwoKey)
            .expect(200)
            .end(function (err, res) {
              if (err) return done(err);
              res.text.should.be.a('string');
              res.text.should.equal(getTestTwoData);
              return done();
            });
    });

    // restify ships with formatters for application/json, text/plain,
    // and application/octet-stream. do not try to use a more specific
    // accept format (e.g. application/pdf)
    // @see http://mcavage.me/node-restify/#Content-Negotiation
    it('should return the correct PDF document (route: GET /document/:key)', function (done) {
      request(url)
            .get('/document/' + getTestThreeKey)
            .expect(200)
            .end(function (err, res) {
              if (err) return done(err);
              return done();
            });
    });

  });


  describe('Test routes: add documents', function () {

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
              filesToCleanup.push(generatedKey);
              response.should.be.an('object');
              should.exist(generatedKey);
              response.message.should.equal('document added');
              request(url)
                    .get('/document/' + generatedKey)
                    .set('accept', 'application/json')
                    .expect(200)
                    .end(function (err, res) {
                      if (err) return done(err);
                      var json = JSON.parse(res.text);
                      json.should.be.an('object');
                      json.testkey1.should.equal('testval');
                      return done();
                    });
            });
    });

    it('should add a JSON document with a defined key (route: PUT /document/:key)', function (done) {
      var key = 'test87654'
        , doc = {'testkey2': 'testval'};
      request(url)
            .put('/document/' + key)
            .set('accept', 'application/json')
            .send(doc)
            .expect(200)
            .end(function (err, res) {
              if (err) return done(err);
              var response = res.body
                , generatedKey = response.key;
              filesToCleanup.push(generatedKey);
              response.should.be.an('object');
              should.exist(generatedKey);
              generatedKey.should.equal(key);
              response.message.should.equal('document added');
              request(url)
                    .get('/document/' + key)
                    .set('accept', 'application/json')
                    .expect(200)
                    .end(function (err, res) {
                      if (err) return done(err);
                      var json = JSON.parse(res.text);
                      json.should.be.an('object');
                      json.testkey2.should.equal('testval');
                      return done();
                    });
            });
    });

    it('should add a text document and return a key (route: PUT /document)', function (done) {
      var doc = 'this is some text';
      request(url)
            .put('/document')
            .send(doc)
            .expect(200)
            .end(function (err, res) {
              if (err) return done(err);
              var response = res.body
                , generatedKey = response.key;
              filesToCleanup.push(generatedKey);
              response.should.be.an('object');
              should.exist(generatedKey);
              response.message.should.equal('document added');
              request(url)
                    .get('/document/' + generatedKey)
                    .expect(200)
                    .end(function (err, res) {
                      if (err) return done(err);
                      res.text.should.equal(doc);
                      return done();
                    });
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