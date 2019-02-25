package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"

	"golang.org/x/net/publicsuffix"
)

// Option is a functional option for configuring the API client
type Option func(*Client) error

// BaseURL allows overriding of API client baseURL for testing
func BaseURL(baseURL string) Option {
	return func(c *Client) error {
		c.baseURL = baseURL
		return nil
	}
}

func Password(password string) Option {
	return func(c *Client) error {
		c.password = password
		return nil
	}
}

func Email(email string) Option {
	return func(c *Client) error {
		c.email = email
		return nil
	}
}

// parseOptions parses the supplied options functions and returns a configured
// *Client instance
func (c *Client) parseOptions(opts ...Option) error {
	// Range over each options function and apply it to our API type to
	// configure it. Options functions are applied in order, with any
	// conflicting options overriding earlier calls.
	for _, option := range opts {
		err := option(c)
		if err != nil {
			return err
		}
	}

	return nil
}

const apiURL = "https://api.localmilkrun.com/v1"
const email = "maeverevels@gmail.com"
const password = "password"

// Client holds information necessary to make a request to your API
type Client struct {
	baseURL    string
	httpClient *http.Client
	email      string
	password   string
}

// New creates a new API client
func New(opts ...Option) (*Client, error) {
	resp, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	client := &Client{
		baseURL: apiURL,
		httpClient: &http.Client{
			Timeout: time.Second * 30,
			Jar:     resp,
		},
		email:    email,
		password: password,
	}

	if err := client.parseOptions(opts...); err != nil {
		return nil, err
	}

	return client, nil
}

func (c *Client) login() error {
	log.Print("Logging In")
	_, err = client.PostForm(authUrl,
		url.Values{"email": {email}, "password": {password}})
	if err != nil {
		log.Fatal(err)
		return err
	}

	log.Print("Logged In")

}

func (c *Client) order() error {
	orderJson := []byte(`{"product_id":1837,"add_quantity":6}`)
	_, err = client.Post(cartContentsUrl, "application/json;charset=utf-8", bytes.NewBuffer(orderJson))
	if err != nil {
		log.Fatal(err)
		return err
	}

	log.Print("Ordered Milk")
}

func (c *Client) checkout() error {
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
}

func (c *Client) logout() error {
	req, _ := http.NewRequest("DELETE", authUrl, nil)
	_, err = http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
		return "", err
	}
	log.Print("Logged out")
}
