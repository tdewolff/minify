const minify = require('./build/Release/obj.target/minify');
console.log(minify.string("text/html", "<span style=\"color:#ff0000;\">A  phrase</span>"));
console.log(minify.file("text/html", "test.html", "test.min.html"));
