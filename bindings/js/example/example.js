import { config, string, file } from '@tdewolff/minify';  // run: npm i @tdewolff/minify

config({'html-keep-document-tags': true})

let s = string("text/html", "<html><span style=\"color:#ff0000;\">A  phrase</span></html>");
console.log(s)

file("text/html", "example.html", "example.min.html");
