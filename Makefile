all: *.go
	mkdir -p build/
	/usr/lib/go-1.8/bin/go build

package:
	dpkg-buildpackage -us -uc
	mv ../sampler_* build/

clean:
	@rm -rf build/
	@rm -f sampler
