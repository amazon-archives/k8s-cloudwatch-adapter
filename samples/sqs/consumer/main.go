package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/awslabs/k8s-cloudwatch-adapter/pkg/aws"
)

func main() {
	// Using the SDK's default configuration, loading additional config
	// and credentials values from the environment variables, shared
	// credentials, and shared configuration files
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic("unable to load SDK config, " + err.Error())
	}

	if cfg.Region == "" {
		cfg.Region = aws.GetLocalRegion()
	}

	fmt.Println("using AWS Region: %s", cfg.Region)

	svc := sqs.NewFromConfig(cfg)

	// Initialize and create a SQS Queue named helloworld if it doesn't exist
	queueName := os.Getenv("QUEUE")
	if queueName == "" {
		queueName = "helloworld"
	}
	fmt.Println("listening to queue:", queueName)

	q, err := svc.GetQueueUrl(context.Background(), &sqs.GetQueueUrlInput{
		QueueName: &queueName,
	})
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

	timeout := int32(20)

	for {
		msg, err := svc.ReceiveMessage(context.Background(), &sqs.ReceiveMessageInput{
			QueueUrl:        q.QueueUrl,
			WaitTimeSeconds: timeout,
		})
		if err != nil {
			fmt.Println("error receiving message from queue:", err)
		} else {
			fmt.Println("message:", msg)
		}
		if len(msg.Messages) > 0 {
			_, err = svc.DeleteMessage(context.Background(), &sqs.DeleteMessageInput{
				QueueUrl:      q.QueueUrl,
				ReceiptHandle: msg.Messages[0].ReceiptHandle,
			})
			if err != nil {
				fmt.Println("error deleting message from queue:", err)
			}
		}
		// Implement some delay here to simulate processing time
		time.Sleep(time.Duration(1000) * time.Millisecond)
	}

}
