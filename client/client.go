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

const (
	apiUrl          = "https://api.localmilkrun.com/v1"
	authUrl         = apiUrl + "/auth"
	cartUrl         = apiUrl + "/shopping_cart"
	cartContentsUrl = cartUrl + "/contents"
	checkoutsUrl    = apiUrl + "/checkouts"
)

func (c *Client) DoMilkrunOrder() (string, error) {
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

	return "success!", nil
}

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

func Transport(transport http.RoundTripper) Option {
	return func(c *Client) error {
		c.httpClient.Transport = transport
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
const default_email = "default_email"
const default_password = "deefault_password"

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
		email:    default_email,
		password: default_password,
	}

	if err := client.parseOptions(opts...); err != nil {
		return nil, err
	}

	return client, nil
}

func (c *Client) login() error {
	log.Print("Logging In")

	_, err := c.httpClient.PostForm(authUrl,
		url.Values{"email": {c.email}, "password": {c.password}})
	if err != nil {
		log.Fatal(err)
		return err
	}

	log.Print("Logged In")
	return nil
}

func (c *Client) order() error {
	orderJson := []byte(`{"product_id":1837,"add_quantity":7}`)
	_, err := c.httpClient.Post(cartContentsUrl, "application/json;charset=utf-8", bytes.NewBuffer(orderJson))
	if err != nil {
		log.Fatal(err)
		return err
	}

	log.Print("Ordered Milk")
	return nil
}

func (c *Client) checkout() error {
	log.Print("Checking out")

	checkoutResponse, err := c.httpClient.Post(checkoutsUrl, "application/json;charset=utf-8", nil)
	if err != nil {
		log.Fatal(err)
		return err
	}

	checkoutJson, err := ioutil.ReadAll(checkoutResponse.Body)
	if err != nil {
		log.Fatal(err)
		return err
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
	finalizeCheckoutResponse, err := c.httpClient.Post(finalizeUrl, "application/json;charset=utf-8", bytes.NewBuffer([]byte(finalizePayload)))
	if err != nil {
		log.Fatal(err)
		return err
	}

	finalizeJson, err := ioutil.ReadAll(finalizeCheckoutResponse.Body)
	if err != nil {
		log.Fatal(err)
		return err
	}
	log.Printf("Finalized! JSON %s", finalizeJson)
	return nil
}

func (c *Client) logout() error {
	req, _ := http.NewRequest("DELETE", authUrl, nil)
	_, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
		return err
	}
	log.Print("Logged out")
	return nil
}
