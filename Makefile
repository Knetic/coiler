default: build
all: package

export GOPATH=$(CURDIR)/
export GOBIN=$(CURDIR)/.temp/

init: clean
	go get ./...

build: init
	go build -o ./.output/coiler .

test:
	go test
	go test -bench=.

clean:
	@rm -rf ./.output/

dist: build test

	export GOOS=linux; \
	export GOARCH=amd64; \
	go build -o ./.output/coiler64 .

	export GOOS=linux; \
	export GOARCH=386; \
	go build -o ./.output/coiler32 .

	export GOOS=darwin; \
	export GOARCH=amd64; \
	go build -o ./.output/coiler_osx .

	export GOOS=windows; \
	export GOARCH=amd64; \
	go build -o ./.output/coiler.exe .



package: dist

ifeq ($(shell which fpm), )
	@echo "FPM is not installed, no packages will be made."
	@echo "https://github.com/jordansissel/fpm"
	@exit 1
endif

ifeq ($(COIL_VERSION), )

	@echo "No 'COIL_VERSION' was specified."
	@echo "Export a 'COIL_VERSION' environment variable to perform a package"
	@exit 1
endif

	fpm \
		--log error \
		-s dir \
		-t deb \
		-v $(COIL_VERSION) \
		-n coiler \
		./.output/coiler64=/usr/local/bin/coiler \
		./docs/coiler.7=/usr/share/man/man7/coiler.7 \
		./autocomplete/coiler=/etc/bash_completion.d/coiler

	fpm \
		--log error \
		-s dir \
		-t deb \
		-v $(COIL_VERSION) \
		-n coiler \
		-a i686 \
		./.output/coiler32=/usr/local/bin/coiler \
		./docs/coiler.7=/usr/share/man/man7/coiler.7 \
		./autocomplete/coiler=/etc/bash_completion.d/coiler

	@mv ./*.deb ./.output/

	fpm \
		--log error \
		-s dir \
		-t rpm \
		-v $(COIL_VERSION) \
		-n coiler \
		./.output/coiler64=/usr/local/bin/coiler \
		./docs/coiler.7=/usr/share/man/man7/coiler.7 \
		./autocomplete/coiler=/etc/bash_completion.d/coiler
	fpm \
		--log error \
		-s dir \
		-t rpm \
		-v $(COIL_VERSION) \
		-n coiler \
		-a i686 \
		./.output/coiler32=/usr/local/bin/coiler \
		./docs/coiler.7=/usr/share/man/man7/coiler.7 \
		./autocomplete/coiler=/etc/bash_completion.d/coiler

	@mv ./*.rpm ./.output/
