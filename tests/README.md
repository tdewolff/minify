# Fuzzing minify

Fuzz testing for `minify` using [go-fuzz](https://github.com/dvyukov/go-fuzz). Pull requests to add more corpora or testers are appreciated.

To run the tests, install `go-fuzz`:

```
GO111MODULE=off go get -u github.com/dvyukov/go-fuzz/go-fuzz github.com/dvyukov/go-fuzz/go-fuzz-build

cd $GOPATH/github.com/tdewolff/minify/tests/html

go-fuzz-build
go-fuzz
```

If restarts is not close to `1/10000`, something is probably wrong. If not finding new corpus for a while, restart the fuzzer.
