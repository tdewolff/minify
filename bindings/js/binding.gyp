{
    "targets": [{
        "target_name": "minify",
        "product_extension": "node",
        "type": "<(library)",
        "cflags": ["-Wall"],
        "sources": ["minify.cc"],
        "libraries": ["../minify.a"],
    }],
}
