package main

import (
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"testing"
)

func TestHandleRequest(t *testing.T) {
	resp, err := HandleRequest(nil, events.APIGatewayProxyRequest{
		Path : "/hello/world",
	})

	if err != nil {
		t.Error("Should not return an error")
		t.Fail()
	}

	if resp.StatusCode != 200 {
		t.Error("Status code should be 200, is", resp.StatusCode)
		t.Fail()
	}

	fmt.Println(resp.Body)
}
