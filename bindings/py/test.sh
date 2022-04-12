go build -buildmode=c-shared -o minify.so
python -c 'import minify; print(minify.string("text/html", "<span style=\"color:#ff0000;\" class=\"text\">A  phrase</span>")); print(minify.file("text/html", "test.html", "test.min.html"));'
