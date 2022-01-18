# Use this image to build the executable
FROM golang:1.16-alpine AS compiler

WORKDIR $GOPATH/src/minify
COPY . .

RUN apk add --update --update-cache --no-cache git ca-certificates && make install

# Final image containing the executable from the previous step
FROM alpine:3

COPY --from=compiler /bin/minify /bin/minify
