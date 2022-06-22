const minify = require('./build/Release/obj.target/minify');

minify.config({'html-keep-document-tags': 5.0})

console.log(minify.string("text/html", "<html><span style=\"color:#ff0000;\">A  phrase</span>"));
