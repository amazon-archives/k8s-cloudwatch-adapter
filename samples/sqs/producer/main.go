package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

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

	// Initialize and create a Service Bus Queue named helloworld if it doesn't exist
	queueName := os.Getenv("QUEUE")
	if queueName == "" {
		queueName = "helloworld"
	}
	message := "Hello SQS."
	fmt.Println("creating queue: ", queueName)

	req, q := svc.CreateQueueRequest(&sqs.CreateQueueInput{
		QueueName: &queueName,
	})
	req.SetContext(context.Background())
	if err := req.Send(); req != nil {
		// handle queue creation error
		fmt.Println("create queue: ", err)
	}

	//https: //stackoverflow.com/a/18158859/697126
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, os.Kill)
	go func() {
		<-signalChan
		os.Exit(1)
	}()

	for i := 1; i < 20000; i++ {
		fmt.Println("sending message ", i)
		req, resp := svc.SendMessageRequest(&sqs.SendMessageInput{
			MessageBody: &message,
			QueueUrl:    q.QueueUrl,
		})
		req.SetContext(context.Background())
		if err := req.Send(); err != nil {
			fmt.Println("Error", err)
			return
		}

		fmt.Println("Success", aws.StringValue(resp.MessageId))
		//		time.Sleep(time.Duration(100) * time.Millisecond)
	}

}
