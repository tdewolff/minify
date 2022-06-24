import { config, string } from '@tdewolff/minify';

config({'html-keep-document-tags': true})

console.log(string("text/html", "<html><span style=\"color:#ff0000;\">A  phrase</span>"));
