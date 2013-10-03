/*global describe:true, before:true, after: true, it:true */

'use strict';

var util = require('util')
  , path = require('path')
  , fs = require('fs')
  , settings = require('yaml-config').readConfig(path.join(__dirname, '..', 'config.yaml'), 'default')
  , should = require('chai').should()
  , request = require('supertest')
  , url = 'http://localhost:' + settings.server.port
  , riak = require('riak-js')
  , riakClient;


describe("test routes", function () {

  before(function (done) {
    
    var server = require('../server');

    // add document to riak
    var servers = settings.riak.servers || ['localhost:8098']
      , client = settings.riak.client + '-test' || 'test'
      , pool = settings.riak.pool || 'test-pool'
      , bucket = settings.riak.bucket || 'default';

    riakClient = riak.getClient({pool: {servers: servers, name: pool, keepAlive: true}, clientId: client});

    riakClient.save(bucket, '23456', {v1: 30, v2: 45}, {} , function(error) {
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

  it('should return not found error (route:  /noroute)', function (done) {
    request(url)
          .get('/get/noroute')
          .set('accept', 'application/json')
          .set('accept-encoding', 'application/gzip')
          .expect(404)
          .end(function (err, res) {
            var response = res.body;
            response.code.should.equal('ResourceNotFound');
            return done();
          });
  });

  it('should return the correct document (route:  /get/:id)', function (done) {
    request(url)
          .get('/get/23456')
          .set('accept', 'application/json')
          .set('accept-encoding', 'application/gzip')
          .expect(200)
          .end(function (err, res) {
            if (err) return done(err);
            var response = JSON.parse(res.text);
            response.should.be.an('object');
            response.v1.should.equal(30);
            response.v2.should.equal(45);
            return done();
          });
  });

  it('should return document not found error (route:  /get/:id)', function (done) {
    request(url)
          .get('/get/abc')
          .set('accept', 'application/json')
          .set('accept-encoding', 'application/gzip')
          .expect(404)
          .end(function (err, res) {
            var response = res.body;
            response.code.should.equal('ResourceNotFound');
            return done();
          });
  });


});