import { minify } from '@tdewolff/minify';

let input = "<html><span style=\"color:#ff0000;\">A  phrase</span>";
let expected = "<html><span style=color:red>A phrase</span>";

async function test() {
    const output = await minify(input, {
        type: 'text/html',
        htmlKeepDocumentTags: true,
    });

    if (output != expected) {
        throw "unexpected output: '" + output + "' instead of '" + expected + "'";
    }
}


test();