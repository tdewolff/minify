#include <node.h>
#include <nan.h>
#include <stdlib.h>
#include <iostream>
#include "minify.h"

namespace minify {
    using v8::FunctionCallbackInfo;
    using v8::Isolate;
    using v8::Local;
    using v8::Object;
    using v8::Array;
    using v8::String;
    using v8::NewStringType;
    using v8::Value;
    using v8::Exception;
    using v8::Context;
    using v8::JSON;

    void config(const FunctionCallbackInfo<Value> &args) {
        Isolate* isolate = args.GetIsolate();

        if (args.Length() < 1) {
            isolate->ThrowException(Exception::TypeError(
                String::NewFromUtf8(isolate, "expected config argument").ToLocalChecked()));
            return;
        }

        if (!args[0]->IsObject()) {
            isolate->ThrowException(Exception::TypeError(
                String::NewFromUtf8(isolate, "config must be an object").ToLocalChecked()));
            return;
        }

        Local<Context> context = Context::New(isolate);
        Local<Object> object = args[0]->ToObject(context).ToLocalChecked();
        Local<Array> properties = object->GetOwnPropertyNames(context).ToLocalChecked();
        uint32_t length = properties->Length();

        std::string cppkeys[length];
        std::string cppvals[length];
        for (uint32_t i = 0; i < length; i++) {
            Local<Value> key = properties->Get(context, i).ToLocalChecked();
            Local<Value> val = object->Get(context, key).ToLocalChecked();

            if (!key->IsString() || (!val->IsString() && !val->IsBoolean() && !val->IsInt32())) {
                isolate->ThrowException(Exception::TypeError(
                    String::NewFromUtf8(isolate, "config must be an object[string: string|boolean|integer]").ToLocalChecked()));
                return;
            }
            if (!val->IsString()) {
                val = JSON::Stringify(context, val).ToLocalChecked();
            }

            std::string cppkey(*String::Utf8Value(isolate, key));
            std::string cppval(*String::Utf8Value(isolate, val));
            cppkeys[i] = cppkey;
            cppvals[i] = cppval;
        }

        const char **keys = (const char **)malloc(length * sizeof(const char *));
        const char **vals = (const char **)malloc(length * sizeof(const char *));
        for (uint32_t i = 0; i < length; i++) {
            keys[i] = cppkeys[i].c_str();
            vals[i] = cppvals[i].c_str();
        }

        char *error = minifyConfig((char **)keys, (char **)vals, (long long)length);
        free(vals);
        free(keys);
        if (error != NULL) {
            isolate->ThrowException(Exception::TypeError(
                String::NewFromUtf8(isolate, error).ToLocalChecked()));
            free(error);
            return;
        }
    }

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

    void init(Local<Object> exports) {
        NODE_SET_METHOD(exports, "config", config);
        NODE_SET_METHOD(exports, "string", string);
        NODE_SET_METHOD(exports, "file", file);
    }

    NAN_MODULE_WORKER_ENABLED(minify, init)
}
