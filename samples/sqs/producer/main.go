package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

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
	fmt.Println("using AWS Region:", cfg.Region)

	svc := sqs.NewFromConfig(cfg)

	// Initialize and create a Service Bus Queue named helloworld if it doesn't exist
	queueName := os.Getenv("QUEUE")
	if queueName == "" {
		queueName = "helloworld"
	}
	message := "Hello SQS."
	fmt.Println("creating queue: ", queueName)

	q, err := svc.CreateQueue(context.Background(), &sqs.CreateQueueInput{
		QueueName: &queueName,
	})
	if err != nil {
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
		result, err := svc.SendMessage(context.Background(), &sqs.SendMessageInput{
			MessageBody: &message,
			QueueUrl:    q.QueueUrl,
		})

		if err != nil {
			fmt.Println("Error", err)
			return
		}

		fmt.Println("Success", *result.MessageId)
		//		time.Sleep(time.Duration(100) * time.Millisecond)
	}

}
