{
    "targets": [{
        "target_name": "minify",
        "product_extension": "node",
        "type": "<(library)",
        "cflags": ["-Wall"],
        "sources": ["minify.c"],
        "libraries": ["../minify.a"],
        "conditions": [
            ['OS=="mac"', {
                # node-gyp 2.x doesn't add this anymore
                # https://github.com/TooTallNate/node-gyp/pull/612
                "xcode_settings": {
                    "CLANG_CXX_LANGUAGE_STANDARD": "c++14",
                    "OTHER_LDFLAGS": ["-undefined dynamic_lookup"],
                },
            }],
            ['OS=="win"', {
                "actions": [{
                    "action_name": "build_go",
                    "message": "Building Go library...",
                    "inputs": ["minify.go", "minify.c"],
                    "outputs": ["minify.a"],
                    "action": ["compile.bat"]
                }],
            }, {
                "actions": [{
                    "action_name": "build_go",
                    "message": "Building Go library...",
                    "inputs": ["minify.go", "minify.c"],
                    "outputs": ["minify.a"],
                    "action": ["make", "compile"]
                }],
            }],
        ],
    }],
}
