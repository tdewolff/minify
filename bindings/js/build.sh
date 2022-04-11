go build -buildmode=c-archive -o minify.a main.go
node-gyp configure
node-gyp build
node test.js
