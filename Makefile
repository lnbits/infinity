lnbits: $(shell find . -name "*.go") client/dist/spa/index.html
	CC=$$(which musl-gcc) go build -ldflags="-s -w -linkmode external -extldflags '-static' -X main.commit=$$(git rev-parse HEAD)" -o lnbits

dev:
	godotenv air -c air.toml

build-dev: $(shell find . -name "*.go")
	CC=$$(which musl-gcc) go build -tags=noembed -ldflags="-s -w -linkmode external -extldflags '-static' -X main.commit=$$(git rev-parse HEAD)" -o lnbits

client/dist/spa/index.html: $(shell find client/src/ -maxdepth 2 -name "*.js" -or -name "*.vue")
	cd client && quasar build --debug
