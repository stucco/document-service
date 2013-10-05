#include <v8.h>
#include <node.h>
#include <node_buffer.h>

#include <stdint.h>
#include <cstring>
#include <vector>

using namespace v8;
using namespace node;

namespace {
  Handle<Value> Decode(const Arguments& args) {
    HandleScope scope;

    if (!Buffer::HasInstance(args[0])) {
      return ThrowException(Exception::TypeError(String::New(
          "Argument should be a Buffer object.")));
    }

    const Local<Object> buffer = args[0]->ToObject();
    size_t length = Buffer::Length(buffer);
    char* data = Buffer::Data(buffer);

    size_t split = 0;
    size_t i = 0;

    Local<String> result = String::Empty();
    while (i < length) {
      if ((data[i] & 0xF0) == 0xF0 && i + 4 <= length) {
        result = String::Concat(result, String::New(data, i));

        // Convert 4-byte UTF-8 to Unicode code point
        uint32_t chr = (((data[i] & 0x07) << 18) | ((data[i+1] & 0x3F) << 12) | ((data[i+2] & 0x3F) << 6) | (data[i+3] & 0x3F)) - 0x10000;

        // Write as surrogate pair
        uint16_t surrogate[2] = {0xD800 | (chr >> 10), 0xDC00 | (chr & 0x3FF)};

        // Concatenate to result
        result = String::Concat(result, String::New(surrogate, 2));

        data += i + 4;
        length -= (i + 4);
        i = 0;
      } else {
        i++;
      }
    }

    return scope.Close(String::Concat(result, String::New(data + split, length - split)));
  }

  void replacement_character(std::vector<char>& vector) {
    vector.push_back(0xEF);
    vector.push_back(0xBB);
    vector.push_back(0xBF);
  }

  Handle<Value> Encode(const Arguments& args) {
    HandleScope scope;

    if (!args[0]->IsString()) {
      return ThrowException(Exception::TypeError(String::New(
          "Argument should be a String.")));
    }

    Local<String> string = args[0]->ToString();
    String::Value int16value(string);
    const uint16_t* data = *int16value;
    const size_t length = string->Length();
    size_t i = 0;

    std::vector<char> accumulator;
    // Reserve enough space for ASCII string
    accumulator.reserve(string->Length());

    while (i < length) {
      const uint16_t chr = data[i];
      if (chr < 0x80) {
        accumulator.push_back((char)chr);
      } else if (chr < 0x800) {
        accumulator.push_back(0xC0 | (chr >> 6));
        accumulator.push_back(0x80 | (chr & 0x3F));
      } else if (chr < 0xD800 || chr >= 0xE000) {
        accumulator.push_back(0xE0 | (chr >> 12));
        accumulator.push_back(0x80 | (chr >> 6 & 0x3F));
        accumulator.push_back(0x80 | (chr & 0x3F));
      } else if (chr >= 0xD800 && chr <= 0xDBFF && i + 1 < length) {
        const uint16_t next = data[++i];
        if (next < 0xDC00 || next > 0xDFFF) {
          replacement_character(accumulator);
          continue;
        }
        const uint32_t unicode = (((chr & 0x3FF) << 10) | (next & 0x3FF)) + 0x10000;
        accumulator.push_back(0xF0 | (unicode >> 18));
        accumulator.push_back(0x80 | (unicode >> 12 & 0x3F));
        accumulator.push_back(0x80 | (unicode >> 6 & 0x3F));
        accumulator.push_back(0x80 | (unicode & 0x3F));
      } else {
        replacement_character(accumulator);
      }
      i++;
    }

    Buffer& slowBuffer = *Buffer::New(accumulator.size());
    char* underlying = Buffer::Data(slowBuffer.handle_);
    memcpy(underlying, &accumulator[0], accumulator.size());

    return scope.Close(slowBuffer.handle_);
  }

  void RegisterModule(Handle<Object> target) {
    target->Set(String::NewSymbol("decode"), FunctionTemplate::New(Decode)->GetFunction());
    target->Set(String::NewSymbol("encode"), FunctionTemplate::New(Encode)->GetFunction());
  }
}

NODE_MODULE(wtf8, RegisterModule);
