package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"github.com/aws/aws-lambda-go/events/apigwevents"
	"github.com/aws/aws-lambda-go/lambda"
	"image"
	"image/color"
	"image/png"
)

func HandleRequest(ctx context.Context, name apigwevents.ApiGatewayProxyRequest) (apigwevents.ApiGatewayProxyResponse, error) {
	f, err := genimage()

	var resp apigwevents.ApiGatewayProxyResponse
	if err != nil {
		resp = apigwevents.ApiGatewayProxyResponse{
			StatusCode: 500,
			Body: err.Error(),
		}
	} else {
		enc := base64.StdEncoding.EncodeToString(f.Bytes())
		resp = apigwevents.ApiGatewayProxyResponse{
			Headers: map[string] string {
				"Content-Type" : "image/png",
			},
			StatusCode: 200,
			Body: enc,
		}
	}
	return resp, nil
}

func genimage() (*bytes.Buffer, error){
	// Create an 100 x 50 image
	img := image.NewRGBA(image.Rect(0, 0, 100, 50))

	// Draw a red dot at (2, 3)
	img.Set(2, 3, color.RGBA{255, 0, 0, 255})
	f := &bytes.Buffer{}

	err := png.Encode(f, img)

	return f, err
}

func main() {
	lambda.Start(HandleRequest)
}
