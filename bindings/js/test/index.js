import { config, string } from '@tdewolff/minify';

config({'html-keep-document-tags': true})

let input = "<html><span style=\"color:#ff0000;\">A  phrase</span>";
let expected = "<html><span style=color:red>A phrase</span>";
let output = string("text/html", input);
if (output != expected) {
    throw "unexpected output: '"+output+"' instead of '"+expected+"'";
}
