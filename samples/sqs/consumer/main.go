package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	util "github.com/awslabs/k8s-cloudwatch-adapter/pkg/aws"
)

func main() {
	// Using the SDK's default configuration, loading additional config
	// and credentials values from the environment variables, shared
	// credentials, and shared configuration files
	cfg := aws.NewConfig()

	if aws.StringValue(cfg.Region) == "" {
		cfg.Region = aws.String(util.GetLocalRegion())
	}
	fmt.Println("using AWS Region:", cfg.Region)

	svc := sqs.New(session.Must(session.NewSession(cfg)))

	// Initialize and create a SQS Queue named helloworld if it doesn't exist
	queueName := os.Getenv("QUEUE")
	if queueName == "" {
		queueName = "helloworld"
	}
	fmt.Println("listening to queue:", queueName)

	req, q := svc.GetQueueUrlRequest(&sqs.GetQueueUrlInput{
		QueueName: &queueName,
	})
	req.SetContext(context.Background())
	if err := req.Send(); req != nil {
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
		req, msg := svc.ReceiveMessageRequest(&sqs.ReceiveMessageInput{
			QueueUrl:        q.QueueUrl,
			WaitTimeSeconds: &timeout,
		})
		req.SetContext(context.Background())
		if err := req.Send(); err == nil {
			fmt.Println("message:", msg)
		} else {
			fmt.Println("error receiving message from queue:", err)
		}
		if len(msg.Messages) > 0 {
			req, _ := svc.DeleteMessageRequest(&sqs.DeleteMessageInput{
				QueueUrl:      q.QueueUrl,
				ReceiptHandle: msg.Messages[0].ReceiptHandle,
			})
			req.SetContext(context.Background())
			if err := req.Send(); err != nil {
				fmt.Println("error deleting message from queue:", err)
			}
		}
		// Implement some delay here to simulate processing time
		time.Sleep(time.Duration(1000) * time.Millisecond)
	}

}
