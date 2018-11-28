package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"image"
	"image/color"
	"image/png"
	"math"
)

func HandleRequest(ctx context.Context, name events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	f, err := genImage()

	var resp events.APIGatewayProxyResponse
	if err != nil {
		resp = events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       err.Error(),
		}
	} else {
		enc := base64.StdEncoding.EncodeToString(f.Bytes())
		resp = events.APIGatewayProxyResponse{
			Headers: map[string]string{
				"Content-Type": "image/png",
			},
			StatusCode:      200,
			Body:            enc,
			IsBase64Encoded: true,
		}
	}
	return resp, nil
}

type Circle struct {
	X, Y, R float64
}

func (c *Circle) Brightness(x, y float64) uint8 {
	var dx, dy float64 = c.X - x, c.Y - y
	d := math.Sqrt(dx*dx+dy*dy) / c.R
	if d > 1 {
		return 0
	} else {
		return 255
	}
}

func genImage() (*bytes.Buffer, error) {
	var w, h int = 280, 240

	m := image.NewRGBA(image.Rect(0, 0, w, h))
	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			c := color.RGBA{
				R: 0,
				G: 0,
				B: 0,
				A: 255,
			}
			m.Set(x, y, c)
		}
	}

	f := &bytes.Buffer{}
	err := png.Encode(f, m)
	return f, err
}

func main() {
	lambda.Start(HandleRequest)
}
