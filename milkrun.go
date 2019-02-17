package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"

	"golang.org/x/net/publicsuffix"
)

const (
	apiUrl          = "https://api.localmilkrun.com/v1"
	authUrl         = apiUrl + "/auth"
	cartUrl         = apiUrl + "/shopping_cart"
	cartContentsUrl = cartUrl + "/contents"
	checkoutsUrl    = apiUrl + "/checkouts"
)

func main() {
	email := os.Getenv("MILKRUN_EMAIL")
	password := os.Getenv("MILKRUN_PASSWORD")

	fmt.Printf("email %s password %s", email, password)
	return
	resp, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		log.Fatal(err)
	}

	client := &http.Client{
		Jar: resp,
	}

	//login

	_, err = client.PostForm(authUrl,
		url.Values{"email": {email}, "password": {password}})
	if err != nil {
		log.Fatal(err)
		return
	}

	//order milk
	orderJson := []byte(`{"product_id":1837,"add_quantity":6}`)
	_, err = client.Post(cartContentsUrl, "application/json;charset=utf-8", bytes.NewBuffer(orderJson))
	if err != nil {
		log.Fatal(err)
		return
	}

	//get checkout

	checkoutResponse, err := client.Post(checkoutsUrl, "application/json;charset=utf-8", nil)
	if err != nil {
		log.Fatal(err)
		return
	}

	checkoutJson, err := ioutil.ReadAll(checkoutResponse.Body)
	if err != nil {
		log.Fatal(err)
		return
	}

	type Checkout struct {
		ID int `json:"id"`
	}

	var checkout Checkout
	json.Unmarshal([]byte(checkoutJson), &checkout)

	//finalize checkout

	finalizePayload := fmt.Sprintf(`{"id": %d}`, checkout.ID)
	finalizeUrl := fmt.Sprintf(apiUrl+"/checkouts/%d/order", checkout.ID)
	finalizeCheckoutResponse, err := client.Post(finalizeUrl, "application/json;charset=utf-8", bytes.NewBuffer([]byte(finalizePayload)))
	if err != nil {
		log.Fatal(err)
		return
	}

	finalizeJson, err := ioutil.ReadAll(finalizeCheckoutResponse.Body)
	if err != nil {
		log.Fatal(err)
		return
	}
	log.Println("JSON %s", finalizeJson)
	//logout

	req, _ := http.NewRequest("DELETE", authUrl, nil)
	_, err = http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
}
