package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"os"

	"github.com/google/uuid"
	flags "github.com/jessevdk/go-flags"
	"github.com/riferrei/srclient"
	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
)

var opts struct {
	Topic             string `short:"t" long:"topic" env:"TOPIC" default:"output" description:"name of the topic"`
	BrokerList        string `short:"b" long:"brokerlist" env:"BROKER_LIST" default:"localhost:9092" description:"Kafka broker list"`
	Mode              string `long:"mode" env:"MODE" default:"producer" description:"Mode: producer or consumer"`
	SchemaID          int    `long:"schemaId" env:"SCHEMA_ID" description:"Schema ID"`
	SchemaRegistryURL string `long:"schemaRegistryURL" env:"SCHEMA_REGISTRY_URL" default:"http://localhost:8081" description:"Schema Registry URL"`
	Message           string `short:"m" long:"msg" env:"MSG" description:"Message"`
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

	if opts.Mode == "producer" {
		// 1) Create the producer as you would normally do using Confluent's Go client
		p, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": opts.BrokerList})
		failOnError(err, "Error to connect to kafka")
		defer p.Close()

		go func() {
			for event := range p.Events() {
				switch ev := event.(type) {
				case *kafka.Message:
					message := ev
					if ev.TopicPartition.Error != nil {
						fmt.Printf("Error delivering the message '%s'\n", message.Key)
					} else {
						fmt.Printf("Message '%s' delivered successfully!\n", message.Key)
					}
				}
			}
		}()

		// 2) Fetch schema definition
		schemaRegistryClient := srclient.CreateSchemaRegistryClient(opts.SchemaRegistryURL)
		schema, err := schemaRegistryClient.GetSchema(opts.SchemaID)
		failOnError(err, "Error getting Avro schema from Confluent Schema Registry")
		schemaIDBytes := make([]byte, 4)
		binary.BigEndian.PutUint32(schemaIDBytes, uint32(schema.ID()))

		// 3) Serialize the record using the schema provided by the client,
		// making sure to include the schema id as part of the record.
		value := []byte(opts.Message)
		native, _, err := schema.Codec().NativeFromTextual(value)
		failOnError(err, "NativeFromTextual error")
		valueBytes, err := schema.Codec().BinaryFromNative(nil, native)
		failOnError(err, "BinaryFromNative error")

		var recordValue []byte
		recordValue = append(recordValue, byte(0))
		recordValue = append(recordValue, schemaIDBytes...)
		recordValue = append(recordValue, valueBytes...)

		key, _ := uuid.NewUUID()
		p.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{
				Topic: &opts.Topic, Partition: kafka.PartitionAny},
			Key: []byte(key.String()), Value: recordValue}, nil)

		p.Flush(15 * 1000)
	} else if opts.Mode == "consumer" {
		// 1) Create the consumer as you would normally do using Confluent's Go client
		c, err := kafka.NewConsumer(&kafka.ConfigMap{
			"bootstrap.servers": opts.BrokerList,
			"group.id":          "myGroup",
			"auto.offset.reset": "earliest",
		})
		failOnError(err, "Error to connect to kafka")
		defer c.Close()
		c.SubscribeTopics([]string{opts.Topic}, nil)

		// 2) Create a instance of the client to retrieve the schemas for each message
		schemaRegistryClient := srclient.CreateSchemaRegistryClient(opts.SchemaRegistryURL)

		for {
			msg, err := c.ReadMessage(-1)
			if err == nil {
				// 3) Recover the schema id from the message and use the
				// client to retrieve the schema from Schema Registry.
				// Then use it to deserialize the record accordingly.
				schemaID := binary.BigEndian.Uint32(msg.Value[1:5])
				schema, err := schemaRegistryClient.GetSchema(int(schemaID))
				failOnError(err, fmt.Sprintf("Error getting the schema with id '%d' %s", schemaID, err))

				native, _, err := schema.Codec().NativeFromBinary(msg.Value[5:])
				failOnError(err, "NativeFromBinary error")

				value, err := schema.Codec().TextualFromNative(nil, native)
				failOnError(err, "TextualFromNative error")
				fmt.Printf("Received: %s\n", string(value))
			} else {
				failOnError(err, "Error consuming the message")
			}
		}

	} else {
		log.Fatalf("Invalid mode")
	}
}
