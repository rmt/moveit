.PHONY: ALL

ALL: myxcb moveit

moveit:
	goimports -w moveit.go
	go build moveit.go
