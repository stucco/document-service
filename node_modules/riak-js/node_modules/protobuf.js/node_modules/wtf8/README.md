# WTF8

A UTF-8 decoder and encoder which can handle characters outside Basic Multilingual Plane for node.js. It's also reasonably fast.

## Why?

V8 doesn't like 4-byte UTF-8 characters. 😞

## Methods

### .decode(buffer)

Decode buffer containing UTF-8 data to JS string.

    var utf8 = require('wtf8');
    utf8.decode(new Buffer([0x68, 0x65, 0x6c, 0x6c, 0x6f]));
    // => 'hello'

### .encode(string)

Encode strings to buffer.

    var utf8 = require('wtf8');
    utf8.encode('hello');
    // => <SlowBuffer 68 65 6c 6c 6f>

## TODO

* stream piping
