build:
	dep ensure
	go build .
aws-build:
	dep ensure
	GOOS=linux go build .
	zip milkrun
test:
	go test 
