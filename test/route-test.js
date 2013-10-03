/*global describe:true, before:true, after: true, it:true */

'use strict';

var util = require('util')
  , path = require('path')
  , fs = require('fs')
  , settings = require('yaml-config').readConfig(path.join(__dirname, '..', 'config.yaml'), 'default')
  , should = require('chai').should()
  , request = require('supertest')
  , url = 'http://localhost:' + settings.server.port
  , Riak = require('riak')
  , riakClient;


describe("test routes", function () {

  before(function (done) {
    
    var server = require('../server');

    // add document to riak

    var servers = settings.riak.servers || ['localhost:8098']
      , client = settings.riak.client + '-test' || 'test'
      , pool = settings.riak.pool || 'test-pool'
      , bucket = settings.riak.bucket || 'default';

    riakClient = new Riak(servers, client, pool);

    riakClient.put(bucket, '23456', {v1: 30, v2: 45}, {} , function(error, response, result) {
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

  it('should return the correct test route (route:  /test)', function (done) {
    request(url)
          .get('/get/23456')
          .set('Accept', 'application/json')
          .expect('Content-Type', 'application/json')
          .expect(200)
          .end(function (err, res) {
            if (err) return done(err);
            var resp = res.body;
            resp.should.be.an('object');
            resp.v1.should.equal(30);
            resp.v2.should.equal(45);
            return done();
          });
  });

});