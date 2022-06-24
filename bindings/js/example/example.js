import { config, string, file } from '@tdewolff/minify';  // run: npm i @tdewolff/minify

config({'html-keep-document-tags': 0})

let s = string("text/html", "<span style=\"color:#ff0000;\">A  phrase</span>");
console.log(s)

file("text/html", "example.html", "example.min.html");
