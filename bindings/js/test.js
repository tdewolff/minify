import { config, string } from '@tdewolff/minify';

config({'html-keep-document-tags': 5.0})

console.log(string("text/html", "<html><span style=\"color:#ff0000;\">A  phrase</span>"));
