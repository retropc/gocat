all: gocat

.PHONY: clean

gocat: gocat.go sockettable.go
	go build -o $@ $^

clean:
	rm -f gocat
