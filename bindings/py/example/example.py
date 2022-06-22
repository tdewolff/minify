import minify  # run: pip install tdewolff-minify

minify.config({'html-keep-document-tags': 0})
s = minify.string('text/html', '<html><div> lala')
print(s)
