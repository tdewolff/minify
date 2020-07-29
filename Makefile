VERSION=`git describe --tags`
FLAGS=-ldflags "-s -w -X 'main.Version=${VERSION}'" -trimpath
ENVS=CGO_ENABLED=0

NAME=minify
CMD=./cmd/minify
TARGETS=linux_amd64 darwin_amd64 freebsd_amd64 netbsd_amd64 openbsd_amd64 windows_amd64

all: install

install:
	@git pull -q --tags
	@${ENVS} go install ${FLAGS} ./cmd/minify
	@. cmd/minify/bash_completion

release:
	@git pull -q --tags
	@rm -rf dist
	@mkdir dist
	@for t in ${TARGETS}; do \
		echo Building $$t...; \
		mkdir dist/$$t; \
		os=$$(echo $$t | cut -f1 -d_); \
		arch=$$(echo $$t | cut -f2 -d_); \
		${ENVS} GOOS=$$os GOARCH=$$arch go build ${FLAGS} -o dist/$$t/${NAME} ${CMD}; \
		\
		cp LICENSE dist/$$t/.; \
		cp cmd/minify/README.md dist/$$t/.; \
		if [ "$$os" == "windows" ]; then \
			mv dist/$$t/${NAME} dist/$$t/${NAME}.exe; \
			zip -jq dist/${NAME}_$$t.zip dist/$$t/*; \
			cd dist; \
			sha256sum ${NAME}_$$t.zip >> checksums.txt; \
			cd ..; \
		else \
			cp cmd/minify/bash_completion dist/$$t/.; \
			cd dist/$$t; \
			tar -czf - * | gzip -9 > ../${NAME}_$$t.tar.gz; \
			cd ..; \
			sha256sum ${NAME}_$$t.tar.gz >> checksums.txt; \
			cd ..; \
		fi; \
		rm -rf dist/$$t; \
	done

.PHONY: install release
