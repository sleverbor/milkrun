package main

import (
	"encoding/base64"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
)

const (
	apiUrl          = "https://api.localmilkrun.com/v1"
	authUrl         = apiUrl + "/auth"
	cartUrl         = apiUrl + "/shopping_cart"
	cartContentsUrl = cartUrl + "/contents"
	checkoutsUrl    = apiUrl + "/checkouts"
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

	c := client.New(Email(email), Password(password), BaseURL(apiUrl))
	err := c.login()
	if err != nil {
		log.Fatal(err)
		return "", err
	}

	err = c.order()
	if err != nil {
		log.Fatal(err)
		return "", err
	}

	err = c.checkout()
	if err != nil {
		log.Fatal(err)
		return "", err
	}

	err = c.logout()
	if err != nil {
		log.Fatal(err)
		return "", err
	}

	return "success", nil
}
