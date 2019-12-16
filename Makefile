build:
	go build .
aws-build:
	GOOS=linux go build .
	zip milkrun.zip milkrun
test:
	go test 
