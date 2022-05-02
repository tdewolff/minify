go get -u all
go build -buildmode=c-archive -o minify.a minify.go

version=$(cat go.mod | grep github.com/tdewolff/minify/v2 | cut -d ' ' -f 2 | cut -b 2-)
sed -i "s/0.0.0/$version/" package.json

export PATH=./node_modules/@mapbox/node-pre-gyp/bin:./node_modules/node-pre-gyp-github/bin:$PATH
node-pre-gyp configure
node-pre-gyp build
node-pre-gyp package
node-pre-gyp-github.js publish --release
