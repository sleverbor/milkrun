package client_test

import (
	"testing"

	"github.com/dnaeon/go-vcr/recorder"
	"github.com/sleverbor/milkrun/client"
)

func TestMilkrun(t *testing.T) {

	r, err := recorder.New("testdata/fixtures/milkrun_order")
	if err != nil {
		t.Fatal(err)
	}
	defer r.Stop()

	email := "test@test.com"

	password := "test"

	c, err := client.New(client.Email(email), client.Password(password), client.Transport(r))
	if err != nil {
		t.Fatal(err)
	}

	if _, err = c.DoMilkrunOrder(); err != nil {
		t.Fatal(err)
	}
}
