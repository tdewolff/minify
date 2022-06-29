{
    "targets": [{
        "target_name": "minify",
        "product_extension": "node",
        "type": "<(library)",
        "cflags": ["-Wall"],
        "sources": ["<(module_root_dir)/minify.c"],
        "libraries": ["<(module_root_dir)/minify.a"],
        "actions": [{
            "action_name": "build_go",
            "message": "Building Go library...",
            "inputs": ["<(module_root_dir)/minify.go", "<(module_root_dir)/minify.c"],
            "outputs": ["<(module_root_dir)/minify.a"],
            "action": ["make", "compile"]
        }],
        "conditions": [
          [
            'OS=="mac"',
            {
              # node-gyp 2.x doesn't add this anymore
              # https://github.com/TooTallNate/node-gyp/pull/612
              "xcode_settings": {
                "CLANG_CXX_LANGUAGE_STANDARD": "c++14",
                "OTHER_LDFLAGS": ["-undefined dynamic_lookup"],
              },
            },
          ]
        ],
    }],
}
