plug/memcached.so: $(shell find . -name '*.go')
	go build -buildmode=plugin -o memcache.so memcache.go