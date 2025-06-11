bin/csi-madrid: bin
	go build -o ./bin/csi-madrid ./cmd/csi-madrid

bin:
	mkdir -p bin

.PHONY: clean
clean:
	rm -rf ./bin

.PHONY: image
image:
	sudo docker build -t csi-madrid:local .
