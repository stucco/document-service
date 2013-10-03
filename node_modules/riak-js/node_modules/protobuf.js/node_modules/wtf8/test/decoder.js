var assert = require("assert");
var utf8 = require("../wtf8");
var helper = require("./common");
var encode = helper.encode;
var test = helper.test;

function rangeStr(begin, end) {
  var str = "";
  var i;
  for (i = begin; i < end; i++) {
    if (i > 0x10000) {
      str += String.fromCharCode(0xD800+((i-0x10000)&0x3FF),0xDC00+(((i-0x10000)>>10)&0x3FF));
    } else {
      str += String.fromCharCode(i);
    }
  }
  return str;
}

function testRange(begin, end) {
  var str = rangeStr(begin, end);
  var buffer = encode(str);
  assert.equal(utf8.decode(buffer), str);
}

test("Decoding US-ASCII characters", testRange.bind(this, 0, 0x80));
test("Decoding 2-byte UTF-8 characters", testRange.bind(this, 0x80, 0x800));
test("Decoding 3-byte UTF-8 characters", function() {
  testRange(0x800, 0xD800);
  testRange(0xE000, 0x10000);
});

test("Decoding 4-byte UTF-8 characters", function() {
  // Quick test
  assert.equal(utf8.decode(new Buffer([0xF0, 0x9D, 0x9B, 0xA2])), '\ud835\udee2');
  testRange(0x10000, 0x110000);
});
