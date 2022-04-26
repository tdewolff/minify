go build -buildmode=c-archive -o minify.a minify.go
node-gyp configure
node-gyp build
node test.js
