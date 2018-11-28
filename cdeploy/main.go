package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/codedeploy"
	lmbda "github.com/aws/aws-sdk-go/service/lambda"

	"github.com/aws/aws-sdk-go/service/rekognition"
	"os"
)

type CodeDeployLifeCycleInput struct {
	DeploymentId                  string `json:"DeploymentId"`
	LifecycleEventHookExecutionId string `json:"LifecycleEventHookExecutionId"`
}

func HandleRequest(ctx context.Context, cdLifeCycle CodeDeployLifeCycleInput) error {
	region := os.Getenv("REGION")
	newLambda := os.Getenv("NewVersion")
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
	})
	if err != nil {
		return err
	}

	cdClient := codedeploy.New(sess)
	lambClient := lmbda.New(sess)

	lambdaInput := &lmbda.InvokeInput{
		FunctionName:   aws.String(newLambda),
		InvocationType: aws.String(lmbda.InvocationTypeRequestResponse),
	}

	resp, err := lambClient.Invoke(lambdaInput)

	var payload events.APIGatewayProxyResponse
	err = json.Unmarshal(resp.Payload, &payload)
	if err != nil {
		sendStatus(cdClient, codedeploy.DeploymentStatusFailed, cdLifeCycle)
		return err
	}

	decoded, err := base64.StdEncoding.DecodeString(string(payload.Body))

	if err != nil {
		sendStatus(cdClient, codedeploy.DeploymentStatusFailed, cdLifeCycle)
		return err
	}

	rekClient := rekognition.New(sess)

	input := &rekognition.DetectLabelsInput{
		Image: &rekognition.Image{
			Bytes: decoded,
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

func sendStatus(cdClient *codedeploy.CodeDeploy, status string, input CodeDeployLifeCycleInput) {
	cdInput := &codedeploy.PutLifecycleEventHookExecutionStatusInput{
		LifecycleEventHookExecutionId: aws.String(input.LifecycleEventHookExecutionId),
		DeploymentId:                  aws.String(input.DeploymentId),
		Status:                        aws.String(status),
	}

	res, err := cdClient.PutLifecycleEventHookExecutionStatus(cdInput)
	if err != nil {
		fmt.Println(err.Error())
	}

	fmt.Println(res.GoString())
}

func main() {
	lambda.Start(HandleRequest)
}
