.PHONY: myxcb

myxcb: myxcb.go
	go fmt myxcb.go
	go build myxcb.go
