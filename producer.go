package main

import (
	"encoding/binary"
	"log"
	"os"
	"strings"

	"github.com/IBM/sarama"
	goavro "github.com/linkedin/goavro/v2"
	srclient "github.com/riferrei/srclient"
	"github.com/spf13/cobra"
)

type producerOptions struct {
	Topic             string
	BrokerList        string
	SchemaID          int
	SchemaRegistryURL string
	Message           string
	MessageFromFile   string
	Auth              string
	TLSclientCertFile string
	TLSclientKeyFile  string
	TLScaCertFile     string
}

func newProducerCommand() *cobra.Command {
	opts := &producerOptions{}

	cmd := &cobra.Command{
		Use:   "producer",
		Short: "Run Kafka Avro producer",
		Run: func(cmd *cobra.Command, args []string) {
			runProducer(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.Topic, "topic", "t", "output", "name of the topic")
	cmd.Flags().StringVarP(&opts.BrokerList, "brokerList", "b", "localhost:9092", "Kafka broker list")
	cmd.Flags().IntVar(&opts.SchemaID, "schemaId", 0, "Schema ID")
	cmd.Flags().StringVar(&opts.SchemaRegistryURL, "schemaRegistryURL", "http://localhost:8081", "Schema Registry URL")
	cmd.Flags().StringVarP(&opts.Message, "msg", "m", "", "Message")
	cmd.Flags().StringVarP(&opts.MessageFromFile, "file", "f", "", "Filename")
	cmd.Flags().StringVarP(&opts.Auth, "auth", "a", "wo", "Auth type")
	cmd.Flags().StringVar(&opts.TLSclientCertFile, "certFile", "./client.cer.pem", "TLS certificate file (in pem format)")
	cmd.Flags().StringVar(&opts.TLSclientKeyFile, "keyFile", "./client.key.pem", "TLS key file (in pem format)")
	cmd.Flags().StringVar(&opts.TLScaCertFile, "caCertFile", "./server.cer.pem", "TLS CA certificate file (in pem format)")

	return cmd
}

func runProducer(opts *producerOptions) {
	brokers := []string{opts.BrokerList}

	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Return.Successes = true
	config.ClientID = "go-kafka-consumer"
	config.Consumer.Return.Errors = true

	if strings.EqualFold(opts.Auth, "TLS") {
		tlsConfig, err := NewTLSConfig(opts.TLSclientCertFile,
			opts.TLSclientKeyFile, opts.TLScaCertFile)
		failOnError(err, "Error creating TLS config")
		config.Net.TLS.Enable = true
		config.Net.TLS.Config = tlsConfig
	}

	schemaRegistryClient := srclient.CreateSchemaRegistryClient(opts.SchemaRegistryURL)

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
}
