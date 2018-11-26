package main

import (
	"context"
	"github.com/aws/aws-lambda-go/events/apigwevents"
	"github.com/aws/aws-lambda-go/lambda"
	"log"
)

func HandleRequest(ctx context.Context, name apigwevents.ApiGatewayProxyRequest) (apigwevents.ApiGatewayProxyResponse, error) {
	log.Println(name.Path)

	resp := apigwevents.ApiGatewayProxyResponse{
		StatusCode: 200,
		Body: "Hello",
	}
	return resp, nil
}

func main() {
	lambda.Start(HandleRequest)
}
