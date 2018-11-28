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
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
	})
	if err != nil {
		return err
	}

	cdClient := codedeploy.New(sess)

	client := http.Client{}
	resp, err := client.Get(exec)
	var b bytes.Buffer

	_, err = io.Copy(&b, resp.Body)
	if err != nil {
		sendStatus(cdClient, codedeploy.DeploymentStatusFailed, cdLifeCycle)
		return err
	}


	rekClient := rekognition.New(sess)

	input := &rekognition.DetectLabelsInput{
		Image: &rekognition.Image{
			Bytes: b.Bytes(),
		},
		MaxLabels: aws.Int64(100),
	}
	out, err := rekClient.DetectLabels(input)

	if err != nil {
		sendStatus(cdClient, codedeploy.DeploymentStatusFailed, cdLifeCycle)
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

	sendStatus(cdClient, codedeploy.DeploymentStatusSucceeded, cdLifeCycle)
	return nil
}

func sendStatus(cdClient *codedeploy.CodeDeploy, status string, input CodeDeployLifeCycleInput)  {
	cdInput := &codedeploy.PutLifecycleEventHookExecutionStatusInput{
		LifecycleEventHookExecutionId: aws.String(input.LifecycleEventHookExecutionId),
		DeploymentId:                  aws.String(input.DeploymentId),
		Status:                        aws.String(status),
	}

	res, err := cdClient.PutLifecycleEventHookExecutionStatus(cdInput)
	fmt.Println(err.Error())
	fmt.Println(res.GoString())
}

func main() {
	lambda.Start(HandleRequest)
}
