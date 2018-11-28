package main

import (
	"bytes"
	"context"
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/codedeploy"
	"github.com/aws/aws-sdk-go/service/rekognition"
	"io"
	"net/http"
	"os"
)

type CodeDeployLifeCycleInput struct {
	DeploymentId string `json:"DeploymentId"`
	LifecycleEventHookExecutionId string `json:"LifecycleEventHookExecutionId"`
}

func HandleRequest(ctx context.Context, cdLifeCycle CodeDeployLifeCycleInput) error {
	uriTemplate := "https://%s.execute-api.%s.amazonaws.com/Prod"
	region := os.Getenv("REGION")
	apiEndpoint := os.Getenv("API_ENDPOINT")
	fmt.Println(os.Getenv("NewVersion"))

	exec := fmt.Sprintf(uriTemplate, apiEndpoint, region)

	client := http.Client{}
	resp, err := client.Get(exec)
	if err != nil {
		return err
	}
	var b bytes.Buffer

	_, err = io.Copy(&b, resp.Body)
	if err != nil {
		return err
	}

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
	})

	rekClient := rekognition.New(sess)

	input := &rekognition.DetectLabelsInput{
		Image: &rekognition.Image{
			Bytes: b.Bytes(),
		},
		MaxLabels: aws.Int64(100),
	}
	out, err := rekClient.DetectLabels(input)

	if err != nil {
		return err
	}

	detections := map[string]bool{}
	for _, label := range out.Labels {
		fmt.Println(*label.Name, *label.Confidence)
		detections[*label.Name] = true
	}

	if _, ok := detections["Moon"]; !ok {
		return fmt.Errorf("Canary could not find circle")
	}

	cdClient := codedeploy.New(sess)

	cdInput := &codedeploy.PutLifecycleEventHookExecutionStatusInput{
		LifecycleEventHookExecutionId: aws.String(cdLifeCycle.LifecycleEventHookExecutionId),
		DeploymentId:                  aws.String(cdLifeCycle.DeploymentId),
		Status:                        aws.String(codedeploy.DeploymentStatusFailed),
	}

	res, err := cdClient.PutLifecycleEventHookExecutionStatus(cdInput)

	if err != nil {
		fmt.Println(err)

		return err
	}
	fmt.Println(res.GoString())
	return nil

}

func main() {
	lambda.Start(HandleRequest)
}
