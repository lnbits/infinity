lnbits: *.go client/dist/spa/index.html
	go build -ldflags="-s -w -X main.commit=$$(git rev-parse HEAD)" -o lnbits

client/dist/spa/index.html: $(shell find client/src/ -maxdepth 2 -name "*.js" -or -name "*.vue")
	cd client && quasar build --debug
