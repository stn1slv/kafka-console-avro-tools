package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	sarama "github.com/Shopify/sarama"
	flags "github.com/jessevdk/go-flags"
	goavro "github.com/linkedin/goavro/v2"
	srclient "github.com/riferrei/srclient"
)

var opts struct {
	Topic             string `short:"t" long:"topic" env:"TOPIC" default:"output" description:"name of the topic"`
	BrokerList        string `short:"b" long:"brokerList" env:"BROKER_LIST" default:"localhost:9092" description:"Kafka broker list"`
	Mode              string `long:"mode" env:"MODE" default:"producer" description:"Mode: producer or consumer"`
	SchemaID          int    `long:"schemaId" env:"SCHEMA_ID" description:"Schema ID"`
	SchemaRegistryURL string `long:"schemaRegistryURL" env:"SCHEMA_REGISTRY_URL" default:"http://localhost:8081" description:"Schema Registry URL"`
	Message           string `short:"m" long:"msg" env:"MSG" description:"Message"`
	MessageFromFile   string `short:"f" long:"file" default:"" env:"FILE" description:"Filename"`
	Group             string `short:"g" long:"group" env:"GROUP" default:"kafka-console-avro-tools" description:"Consumer group"`
	Auth              string `short:"a" long:"auth" env:"AUTH" default:"wo" description:"Auth type"`
	TLSCertsDir       string `short:"x" long:"certDir" env:"TLS_CERTS_DIR" default:"./" description:"Directory with TLS Certificates"`
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func main() {
	if _, err := flags.Parse(&opts); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

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

	if opts.Mode == "producer" {
		// 1) Create the producer as you would normally do using Confluent's Go client
		p, err := sarama.NewSyncProducer(brokers, config)
		failOnError(err, "Error connecting to kafka")
		defer p.Close()

		// 2) Fetch schema definition
		schema, err := schemaRegistryClient.GetSchema(opts.SchemaID)
		failOnError(err, "Error getting Avro schema from Confluent Schema Registry")
		schemaIDBytes := make([]byte, 4)
		binary.BigEndian.PutUint32(schemaIDBytes, uint32(schema.ID()))

		// 3) Serialize the record using the schema provided by the client,
		// making sure to include the schema id as part of the record.
		var value []byte
		if opts.MessageFromFile != "" {
			// Read message payload from a file
			value, err = os.ReadFile(opts.MessageFromFile)
			failOnError(err, "Error when opening file")
			log.Printf("Reading message payload from a file: %s", opts.MessageFromFile)
		} else {
			// Read message payload from input parameter
			value = []byte(opts.Message)
		}

		codec, err := goavro.NewCodecForStandardJSONFull(schema.Schema())
		failOnError(err, "NewCodecForStandardJSON error")
		native, _, err := codec.NativeFromTextual(value)
		failOnError(err, "NativeFromTextual error")
		valueBytes, err := codec.BinaryFromNative(nil, native)
		failOnError(err, "BinaryFromNative error")

		var recordValue []byte
		recordValue = append(recordValue, byte(0))
		recordValue = append(recordValue, schemaIDBytes...)
		recordValue = append(recordValue, valueBytes...)

		// key, _ := uuid.NewUUID()
		partition, offset, err := p.SendMessage(&sarama.ProducerMessage{
			Topic: opts.Topic,
			Value: sarama.ByteEncoder(recordValue),
		})
		failOnError(err, "Failed to send message")

		log.Printf("Wrote message at partition: %d, offset: %d\n", partition, offset)
	} else if opts.Mode == "consumer" {
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
	} else {
		log.Fatalf("Invalid mode. Please specify producer or consumer")
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

// Configuration for TLS Authentication
func NewTLSConfig(clientCertFile, clientKeyFile, caCertFile string) (*tls.Config, error) {
	tlsConfig := tls.Config{}

	// Load client cert
	cert, err := tls.LoadX509KeyPair(clientCertFile, clientKeyFile)
	if err != nil {
		return &tlsConfig, err
	}
	tlsConfig.Certificates = []tls.Certificate{cert}

	// Load CA cert
	caCert, err := ioutil.ReadFile(caCertFile)
	if err != nil {
		return &tlsConfig, err
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	tlsConfig.RootCAs = caCertPool

	tlsConfig.BuildNameToCertificate()
	return &tlsConfig, err
}
