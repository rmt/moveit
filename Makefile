.PHONY: ALL

ALL:
	goimports -w moveit.go
	go build moveit.go
	strip moveit
