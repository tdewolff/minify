import minify  # run: pip install tdewolff-minify

s = minify.string('text/html', '<span style="color:#ff0000;" class="text">Some  text</span>')
print(s)  # <span style=color:red class=text>Some text</span>

minify.file('text/html', 'test.html', 'test.min.html')  # creates test.min.html from test.html
