GO_SOURCES = $(shell find . -name '*.go')

## Development
go.fmt:
	go mod tidy
	go fmt ./...

go.vet:
	CGO_LDFLAGS_ALLOW=-Wl,-z,now go vet ./...

go.unit:
	CGO_LDFLAGS_ALLOW=-Wl,-z,now go test -v ./pkg/... -tags=libsqlite3

## k8s-dqlite tests
go.k8s-dqlite.test:
	CGO_LDFLAGS_ALLOW=-Wl,-z,now go test -tags=libsqlite3 -v ./test/k8s-dqlite

go.k8s-dqlite.test.dqlite:
	CGO_LDFLAGS_ALLOW=-Wl,-z,now go test -tags=dqlite,libsqlite3 -v ./test/k8s-dqlite

go.k8s-dqlite.bench:
	CGO_LDFLAGS_ALLOW=-Wl,-z,now go test -tags=libsqlite3 -v ./test/k8s-dqlite -run "^$$" -bench "Benchmark" -benchmem

go.k8s-dqlite.bench.dqlite:
	CGO_LDFLAGS_ALLOW=-Wl,-z,now go test -tags=dqlite,libsqlite3 -v ./test/k8s-dqlite -run "^$$" -bench "Benchmark" -benchmem

## Static Builds
static: bin/static/k8s bin/static/k8sd bin/static/k8s-dqlite bin/static/k8s-apiserver-proxy bin/static/dqlite

deps/static/lib/libdqlite.a: hack/compile-static-dqlite.sh
	./hack/compile-static-dqlite.sh

bin/static/k8s: deps/static/lib/libdqlite.a $(GO_SOURCES)
	mkdir -p bin/static
	env \
		PATH="${PATH}:${PWD}/deps/static/musl/bin" \
		CGO_CFLAGS="-I${PWD}/deps/static/include" \
		CGO_LDFLAGS="-L${PWD}/deps/static/lib -luv -lraft -ldqlite -llz4 -lsqlite3" \
		CGO_LDFLAGS_ALLOW="-Wl,-z,now" \
		CC="musl-gcc" \
		go build \
			-tags=dqlite,libsqlite3 \
			-o bin/static/k8s \
			-ldflags '-s -w -linkmode "external" -extldflags "-static"' \
			cmd/main.go

bin/static/k8s-apiserver-proxy: bin/static/k8s
	cd bin/static && ln -sf k8s k8s-apiserver-proxy

bin/static/k8s-dqlite: bin/static/k8s
	cd bin/static && ln -sf k8s k8s-dqlite

bin/static/k8sd: bin/static/k8s
	cd bin/static && ln -sf k8s k8sd

bin/static/dqlite: deps/static/lib/libdqlite.a
	mkdir -p bin/static
	rm -rf build/go-dqlite
	git clone https://github.com/canonical/go-dqlite --depth 1 -b v1.20.0 build/go-dqlite
	cd build/go-dqlite && env \
		PATH="${PATH}:${PWD}/deps/static/musl/bin" \
		CGO_CFLAGS="-I${PWD}/deps/static/include" \
		CGO_LDFLAGS="-L${PWD}/deps/static/lib -luv -lraft -ldqlite -llz4 -lsqlite3" \
		CGO_LDFLAGS_ALLOW="-Wl,-z,now" \
		CC="musl-gcc" \
		go build \
			-tags=dqlite,libsqlite3 \
			-ldflags '-s -w -linkmode "external" -extldflags "-static"' \
			-o ${PWD}/bin/static/dqlite \
			github.com/canonical/go-dqlite/cmd/dqlite

## Dynamic Builds
dynamic: bin/dynamic/lib/libdqlite.so bin/dynamic/k8s bin/dynamic/k8sd bin/dynamic/k8s-dqlite bin/dynamic/k8s-apiserver-proxy bin/dynamic/dqlite

bin/dynamic/lib/libdqlite.so: deps/dynamic/lib/libdqlite.so
	mkdir -p bin/dynamic/lib
	cp deps/dynamic/lib/*.so* bin/dynamic/lib

deps/dynamic/lib/libdqlite.so:
	./hack/compile-dynamic-dqlite.sh

bin/dynamic/k8s: deps/dynamic/lib/libdqlite.so
	mkdir -p bin/dynamic
	env \
		CGO_CFLAGS="-I${PWD}/deps/dynamic/include" \
		CGO_LDFLAGS="-L${PWD}/deps/dynamic/lib -lraft -luv -llz4 -lsqlite3" \
		CGO_LDFLAGS_ALLOW="-Wl,-z,now" \
		go build \
			-tags=dqlite,libsqlite3 \
			-o bin/dynamic/k8s \
			-ldflags '-s -w' \
			cmd/main.go

bin/dynamic/k8s-apiserver-proxy: bin/dynamic/k8s
	cd bin/dynamic && ln -s k8s k8s-apiserver-proxy

bin/dynamic/k8s-dqlite: bin/dynamic/k8s
	cd bin/dynamic && ln -s k8s k8s-dqlite

bin/dynamic/k8sd: bin/dynamic/k8s
	cd bin/dynamic && ln -s k8s k8sd

bin/dynamic/dqlite: deps/dynamic/lib/libdqlite.so
	mkdir -p bin/dynamic
	rm -rf build/go-dqlite
	git clone https://github.com/canonical/go-dqlite --depth 1 -b v1.20.0 build/go-dqlite
	cd build/go-dqlite && env \
		CGO_CFLAGS="-I${PWD}/deps/dynamic/include" \
		CGO_LDFLAGS="-L${PWD}/deps/dynamic/lib -luv -lraft -ldqlite -llz4 -lsqlite3" \
		CGO_LDFLAGS_ALLOW="-Wl,-z,now" \
		go build \
			-tags=dqlite,libsqlite3 \
			-ldflags '-s -w' \
			-o ${PWD}/bin/dynamic/dqlite \
			./cmd/dqlite

## Cleanup
clean:
	rm -rf bin build deps
