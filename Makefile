lnbits: $(shell find . -name "*.go") client/dist/spa/index.html
	CC=$$(which musl-gcc) go build -ldflags="-s -w -linkmode external -extldflags '-static' -X main.commit=$$(git rev-parse HEAD)" -o lnbits

client/dist/spa/index.html: $(shell find client/src/ -maxdepth 2 -name "*.js" -or -name "*.vue")
	cd client && quasar build --debug
