package main

import (
	"encoding/base64"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
	milkrun "github.com/sleverbor/milkrun/client"
)

var encrypted string = os.Getenv("MILKRUN_PASSWORD")
var decrypted string

func init() {
	kmsClient := kms.New(session.New())
	decodedBytes, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		panic(err)
	}
	input := &kms.DecryptInput{
		CiphertextBlob: decodedBytes,
	}
	response, err := kmsClient.Decrypt(input)
	if err != nil {
		panic(err)
	}
	// Plaintext is a byte array, so convert to string
	decrypted = string(response.Plaintext[:])
}

func main() {
	lambda.Start(HandleRequest)
}

func HandleRequest() (string, error) {
	log.Print("Starting Milkrun")
	email := os.Getenv("MILKRUN_EMAIL")
	password := decrypted

	c, err := milkrun.New(milkrun.Email(email), milkrun.Password(password))
	if err != nil {
		panic(err)
	}

	if _, err = c.DoMilkrunOrder(); err != nil {
		panic(err)
	}

	return "success", nil
}
