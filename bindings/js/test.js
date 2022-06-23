import { config, string, file } from '@tdewolff/minify';

config({'html-keep-document-tags': "a"})

console.log(string("text/html", "<html><span style=\"color:#ff0000;\">A  phrase</span>"));

file("text/html", "test.html", "test.min.html");
