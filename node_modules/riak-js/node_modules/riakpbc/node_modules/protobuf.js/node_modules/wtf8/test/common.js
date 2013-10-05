var passed = 0;
var failed = 0;

exports.test = function test(name, test) {
  try {
    test();
    passed += 1;
    console.log(name + " passed.");
  } catch (err) {
    console.log("\x1B[1;31m" + name + " failed!\x1B[0m");
    console.log(err.stack + "\n");
    failed += 1;
  }
}

exports.toHex = function toHex(buffer) {
  var i, parts = [];
  for(i = 0; i < buffer.length; i++) {
    parts.push(buffer[i].toString(16))
  }
  return parts.join(" ")
}

// Safe but slow encoding
exports.encode = function encode(string) {
  var utf8 = unescape(encodeURIComponent(string));
  var len = utf8.length;
  var buffer = new Buffer(utf8.length);
  while(len--) {
    buffer[len] = utf8.charCodeAt(len);
  }
  return buffer;
}

process.on('exit', function() {
  console.log("\n\x1B[1;37mTest runner completed. " + passed + " passed, " + failed + " failed.\x1B[0m");
  if (failed > 0) process.exit(1);
});
