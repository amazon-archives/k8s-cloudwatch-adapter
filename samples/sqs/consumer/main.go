package main

import (
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/awslabs/k8s-cloudwatch-adapter/pkg/aws"
)

func main() {
	// Using the SDK's default configuration, loading additional config
	// and credentials values from the environment variables, shared
	// credentials, and shared configuration files
	cfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		panic("unable to load SDK config, " + err.Error())
	}

	if cfg.Region == "" {
		cfg.Region = aws.GetLocalRegion()
	}
	fmt.Println("using AWS Region:", cfg.Region)

	svc := sqs.New(cfg)

	// Initialize and create a SQS Queue named helloworld if it doesn't exist
	queueName := os.Getenv("QUEUE")
	if queueName == "" {
		queueName = "helloworld"
	}
	fmt.Println("listening to queue:", queueName)

	q, err := svc.GetQueueUrlRequest(&sqs.GetQueueUrlInput{
		QueueName: &queueName,
	}).Send()
	if err != nil {
		// handle queue creation error
		fmt.Println("cannot get queue:", err)
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, os.Kill)
	go func() {
		<-signalChan
		os.Exit(1)
	}()

	timeout := int64(20)

	for {
		msg, err := svc.ReceiveMessageRequest(&sqs.ReceiveMessageInput{
			QueueUrl:        q.QueueUrl,
			WaitTimeSeconds: &timeout,
		}).Send()
		if err != nil {
			fmt.Println("error receiving message from queue:", err)
		} else {
			fmt.Println("message:", msg)
		}
		_, err = svc.DeleteMessageRequest(&sqs.DeleteMessageInput{
			QueueUrl:      q.QueueUrl,
			ReceiptHandle: msg.Messages[0].ReceiptHandle,
		}).Send()
		if err != nil {
			fmt.Println("error deleting message from queue:", err)
		}
		// Implement some delay here to simulate processing time
		time.Sleep(time.Duration(1000) * time.Millisecond)
	}

}
