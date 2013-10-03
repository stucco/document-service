var assert = require("assert");
var utf8 = require("../wtf8");
var helper = require("./common");
var test = helper.test;
var toHex = helper.toHex;
var encode = helper.encode;

function testRange(begin, end) {
  var str = "", c, expected, result;
  for(var i = begin; i < end; i++) {
    if (i >= 0x10000) {
      str += c = (String.fromCharCode(0xD800+(((i-0x10000) >> 10)&0x3FF),0xDC00+(((i-0x10000))&0x3FF)));
    } else {
      str += c = String.fromCharCode(i);
    }

    expected = encode(c);
    result = utf8.encode(c);
    assert.equal(toHex(result), toHex(expected), "encode 0x" + i.toString(16) + " (was: " + toHex(result) + ", expected: " + toHex(expected) + ")");    
  }

  assert.equal(toHex(utf8.encode(str)), toHex(encode(str)));
}

test("Encoding US-ASCII characters", testRange.bind(this, 0, 128));
test("Encoding 2-byte UTF-8 characters", testRange.bind(this, 0x80, 0x800));
test("Encoding 3-byte UTF-8 characters", testRange.bind(this, 0x800, 0x1000));
test("Encoding 4-byte UTF-8 characters", function() {
  // Quick test
  assert.equal(toHex(utf8.encode('\ud835\udee2')), toHex(new Buffer([0xF0, 0x9D, 0x9B, 0xA2])));
  testRange(0x10000, 0x110000);
});

test("Encoding invalid characters", function() {
  assert.equal(toHex(utf8.encode('\ud800hello')), toHex(encode("\ufeffhello")));
  assert.equal(toHex(utf8.encode('hello\udc00')), toHex(encode("hello\ufeff")));
});

test("Encoding mixed string", function() {
  var str = "Testing real string with 4-byte chars like \ud840\udc00 in the middle.";
  assert.equal(toHex(utf8.encode(str)), toHex(encode(str)));
})
