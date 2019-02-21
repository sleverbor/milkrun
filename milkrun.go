package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
	"golang.org/x/net/publicsuffix"
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

	resp, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		log.Fatal(err)
		return "", err
	}

	client := &http.Client{
		Jar: resp,
	}

	//login
	log.Print("Logging In")
	_, err = client.PostForm(authUrl,
		url.Values{"email": {email}, "password": {password}})
	if err != nil {
		log.Fatal(err)
		return "", err
	}

	log.Print("Logged In")

	//order milk
	orderJson := []byte(`{"product_id":1837,"add_quantity":6}`)
	_, err = client.Post(cartContentsUrl, "application/json;charset=utf-8", bytes.NewBuffer(orderJson))
	if err != nil {
		log.Fatal(err)
		return "", err
	}

	log.Print("Ordered Milk")
	//get checkout

	log.Print("Checking out")
	checkoutResponse, err := client.Post(checkoutsUrl, "application/json;charset=utf-8", nil)
	if err != nil {
		log.Fatal(err)
		return "", err
	}

	checkoutJson, err := ioutil.ReadAll(checkoutResponse.Body)
	if err != nil {
		log.Fatal(err)
		return "", err
	}

	type Checkout struct {
		ID int `json:"id"`
	}

	var checkout Checkout
	json.Unmarshal([]byte(checkoutJson), &checkout)

	//finalize checkout

	log.Print("Finalizing Checkout")
	finalizePayload := fmt.Sprintf(`{"id": %d}`, checkout.ID)
	finalizeUrl := fmt.Sprintf(apiUrl+"/checkouts/%d/order", checkout.ID)
	finalizeCheckoutResponse, err := client.Post(finalizeUrl, "application/json;charset=utf-8", bytes.NewBuffer([]byte(finalizePayload)))
	if err != nil {
		log.Fatal(err)
		return "", err
	}

	finalizeJson, err := ioutil.ReadAll(finalizeCheckoutResponse.Body)
	if err != nil {
		log.Fatal(err)
		return "", err
	}
	log.Printf("Finalized! JSON %s", finalizeJson)
	//logout

	req, _ := http.NewRequest("DELETE", authUrl, nil)
	_, err = http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
		return "", err
	}
	log.Print("Logged out")
	return "success", nil
}
