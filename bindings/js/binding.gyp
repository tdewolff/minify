{
    "targets": [{
        "target_name": "minify",
        "product_extension": "node",
        "include_dirs": ["<!(node -e \"require('nan')\")"],
        "type": "<(library)",
        "cflags": ["-Wall"],
        "sources": ["minify.cc"],
        "libraries": ["../minify.a"],
    }],
}
