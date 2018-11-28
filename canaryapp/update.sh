#!/usr/bin/env bash

export GO111MODULE=on
GOOS=linux go build
zip canaryapp.zip canaryapp
aws cloudformation package --template canary-app.yml --s3-bucket canary-app-us-west-2-326750834372-pipe --output-template canary-app-export.yml
aws --region us-west-2 cloudformation deploy --template-file /Users/seamyers/go/src/github.com/iph/yocal/canaryapp/canary-app-export.yml --stack-name canary-stack --parameter-overrides AppName=canary-app ApplicationEndpoint=4n6igpgd86 --capabilities CAPABILITY_IAM
rm canaryapp.zip
rm canary-app-export.yml
rm canaryapp
