{
    "targets": [{
        "target_name": "minify",
        "product_extension": "node",
        "type": "loadable_module",
        "cflags": ["-Wall"],
        "sources": ["minify.c"],
        "libraries": ["../minify.a"],
        "actions": [{
            "action_name": "build_go",
            "message": "Building Go library...",
            "inputs": ["minify.go", "minify.c"],
            "outputs": ["minify.a"],
            "action": ["make", "compile"]
        }],
    }],
}
