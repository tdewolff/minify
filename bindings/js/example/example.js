const minify = require('@tdewolff/minify');  // run: npm i @tdewolff/minify

minify.config({'html-keep-document-tags': 0})

s = minify.string("text/html", "<span style=\"color:#ff0000;\">A  phrase</span>");
console.log(s)

minify.file("text/html", "example.html", "example.min.html");
