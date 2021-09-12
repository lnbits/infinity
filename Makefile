lnbits: $(shell find . -name "*.go")
	go build -ldflags="-s -w" -o lnbits
