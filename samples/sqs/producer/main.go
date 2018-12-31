package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

func main() {
	// Using the SDK's default configuration, loading additional config
	// and credentials values from the environment variables, shared
	// credentials, and shared configuration files
	cfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		panic("unable to load SDK config, " + err.Error())
	}

	svc := sqs.New(cfg)

	// Initialize and create a Service Bus Queue named helloworld if it doesn't exist
	queueName := os.Getenv("QUEUE")
	if queueName == "" {
		queueName = "helloworld"
	}
	message := "Hello SQS."
	fmt.Println("creating queue: ", queueName)

	q, err := svc.CreateQueueRequest(&sqs.CreateQueueInput{
		QueueName: &queueName,
	}).Send()
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
		result, err := svc.SendMessageRequest(&sqs.SendMessageInput{
			MessageBody: &message,
			QueueUrl:    q.QueueUrl,
		}).Send()

		if err != nil {
			fmt.Println("Error", err)
			return
		}

		fmt.Println("Success", *result.MessageId)
		//		time.Sleep(time.Duration(100) * time.Millisecond)
	}

}
