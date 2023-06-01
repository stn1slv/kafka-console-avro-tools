package main

import (
	"context"
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/Shopify/sarama"
	srclient "github.com/riferrei/srclient"
	"github.com/spf13/cobra"
)

type consumerOptions struct {
	Topic             string
	BrokerList        string
	Group             string
	Auth              string
	TLSCertsDir       string
	SchemaRegistryURL string
}

func newConsumerCommand() *cobra.Command {
	opts := &consumerOptions{}

	cmd := &cobra.Command{
		Use:   "consumer",
		Short: "Run Kafka Avro consumer",
		Run: func(cmd *cobra.Command, args []string) {
			runConsumer(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.Topic, "topic", "t", "output", "name of the topic")
	cmd.Flags().StringVarP(&opts.BrokerList, "brokerList", "b", "localhost:9092", "Kafka broker list")
	cmd.Flags().StringVarP(&opts.Group, "group", "g", "kafka-console-avro-tools", "Consumer group")
	cmd.Flags().StringVarP(&opts.Auth, "auth", "a", "wo", "Auth type")
	cmd.Flags().StringVarP(&opts.TLSCertsDir, "certDir", "x", "./", "Directory with TLS Certificates")
	cmd.Flags().StringVar(&opts.SchemaRegistryURL, "schemaRegistryURL", "http://localhost:8081", "Schema Registry URL")

	return cmd
}

func runConsumer(opts *consumerOptions) {
	brokers := []string{opts.BrokerList}

	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Return.Successes = true
	config.ClientID = "go-kafka-consumer"
	config.Consumer.Return.Errors = true

	if opts.Auth == "TLS" {
		tlsConfig, err := NewTLSConfig(opts.TLSCertsDir+"client.cer.pem",
			opts.TLSCertsDir+"client.key.pem", opts.TLSCertsDir+"server.cer.pem")
		failOnError(err, "Error creating TLS config")
		config.Net.TLS.Enable = true
		config.Net.TLS.Config = tlsConfig
	}

	schemaRegistryClient := srclient.CreateSchemaRegistryClient(opts.SchemaRegistryURL)

	/**
	 * Setup a new Sarama consumer group
	 */
	consumer := Consumer{
		ready:    make(chan bool),
		srClient: schemaRegistryClient,
	}

	ctx, cancel := context.WithCancel(context.Background())
	client, err := sarama.NewConsumerGroup(brokers, opts.Group, config)
	if err != nil {
		log.Panicf("Error creating consumer group client: %v", err)
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			// `Consume` should be called inside an infinite loop, when a
			// server-side rebalance happens, the consumer session will need to be
			// recreated to get the new claims
			if err := client.Consume(ctx, []string{opts.Topic}, &consumer); err != nil {
				log.Panicf("Error from consumer: %v", err)
			}
			// check if context was cancelled, signaling that the consumer should stop
			if ctx.Err() != nil {
				return
			}
			consumer.ready = make(chan bool)
		}
	}()

	<-consumer.ready // Await till the consumer has been set up
	log.Println("Consumer up and running!...")

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-ctx.Done():
		log.Println("terminating: context cancelled")
	case <-sigterm:
		log.Println("terminating: via signal")
	}
	cancel()
	wg.Wait()
	if err = client.Close(); err != nil {
		log.Panicf("Error closing client: %v", err)
	}
}

// Consumer represents a Sarama consumer group consumer
type Consumer struct {
	ready    chan bool
	srClient *srclient.SchemaRegistryClient
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (consumer *Consumer) Setup(sarama.ConsumerGroupSession) error {
	// Mark the consumer as ready
	close(consumer.ready)
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (consumer *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
func (consumer *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	// NOTE:
	// Do not move the code below to a goroutine.
	// The `ConsumeClaim` itself is called within a goroutine, see:
	// https://github.com/Shopify/sarama/blob/main/consumer_group.go#L27-L29
	for message := range claim.Messages() {

		schemaID := binary.BigEndian.Uint32(message.Value[1:5])
		schema, err := consumer.srClient.GetSchema(int(schemaID))
		failOnError(err, fmt.Sprintf("Error getting the schema with id '%d' %s", schemaID, err))
		native, _, err := schema.Codec().NativeFromBinary(message.Value[5:])
		failOnError(err, "NativeFromBinary error")

		value, err := schema.Codec().TextualFromNative(nil, native)
		failOnError(err, "TextualFromNative error")
		fmt.Printf("Received: %s\n", string(value))

		session.MarkMessage(message, "")
	}

	return nil
}
