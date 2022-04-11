#include <node.h>
#include <stdlib.h>
#include "minify.h"

namespace minify {
    using v8::FunctionCallbackInfo;
    using v8::Isolate;
    using v8::Local;
    using v8::Object;
    using v8::String;
    using v8::NewStringType;
    using v8::Value;
    using v8::Exception;

    void string(const FunctionCallbackInfo<Value> &args) {
        Isolate* isolate = args.GetIsolate();

        if (args.Length() < 2) {
            isolate->ThrowException(Exception::TypeError(
                String::NewFromUtf8(isolate, "expected mediatype and input arguments").ToLocalChecked()));
            return;
        }

        if (!args[0]->IsString() || !args[1]->IsString()) {
            isolate->ThrowException(Exception::TypeError(
                String::NewFromUtf8(isolate, "expected mediatype and input arguments").ToLocalChecked()));
            return;
        }

        std::string cppmediatype(*String::Utf8Value(isolate, args[0]));
        std::string cppinput(*String::Utf8Value(isolate, args[1]));

        const char *mediatype = cppmediatype.c_str();
        const char *input = cppinput.c_str();
        long long input_length = strlen(input); // not including trailing NULL-byte

        long long output_length; // not including trailing NULL-byte
        char *output = (char *)malloc(input_length);
        char *error = minifyString((char *)mediatype, (char *)input, (long long)input_length, output, &output_length);
        if (error != NULL) {
            isolate->ThrowException(Exception::TypeError(
                String::NewFromUtf8(isolate, error).ToLocalChecked()));
            free(error);
            return;
        }

        args.GetReturnValue().Set(String::NewFromUtf8(isolate, output, NewStringType::kNormal, (int)output_length).ToLocalChecked());
        free(output);
    }

    void file(const FunctionCallbackInfo<Value> &args) {
        Isolate* isolate = args.GetIsolate();

        if (args.Length() < 3) {
            isolate->ThrowException(Exception::TypeError(
                String::NewFromUtf8(isolate, "expected mediatype, input, and output arguments").ToLocalChecked()));
            return;
        }

        if (!args[0]->IsString() || !args[1]->IsString() || !args[2]->IsString()) {
            isolate->ThrowException(Exception::TypeError(
                String::NewFromUtf8(isolate, "expected mediatype, input, and output arguments").ToLocalChecked()));
            return;
        }

        std::string cppmediatype(*String::Utf8Value(isolate, args[0]));
        std::string cppinput(*String::Utf8Value(isolate, args[1]));
        std::string cppoutput(*String::Utf8Value(isolate, args[2]));

        const char *mediatype = cppmediatype.c_str();
        const char *input = cppinput.c_str();
        const char *output = cppoutput.c_str();

        char *error = minifyFile((char *)mediatype, (char *)input, (char *)output);
        if (error != NULL) {
            isolate->ThrowException(Exception::TypeError(
                String::NewFromUtf8(isolate, error).ToLocalChecked()));
            free(error);
            return;
        }
    }

    void init(Local<Object> exports, Local<Value> module, void *context) {
        NODE_SET_METHOD(exports, "string", string);
        NODE_SET_METHOD(exports, "file", file);
    }

    NODE_MODULE(minify, init)
}
