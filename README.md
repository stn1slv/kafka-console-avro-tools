# Kafka Avro Producer and Consumer

This application demonstrates how to use Kafka with Avro serialization and deserialization. It includes a producer and a consumer that can send and receive Avro-encoded messages using the Confluent Schema Registry.

## Prerequisites

- Go 1.16 or later
- Kafka cluster with Confluent Schema Registry

## Dependencies

- [Shopify/sarama](https://github.com/Shopify/sarama)
- [linkedin/goavro](https://github.com/linkedin/goavro)
- [riferrei/srclient](https://github.com/riferrei/srclient)
- [spf13/cobra](https://github.com/spf13/cobra)

## Usage

### Build

To build the application, run the following command:

```sh
go build -o kafka-avro
```
### Producer

To run the producer, use the following command:
```sh
./kafka-avro producer [flags]
```
Flags:
- `-t, --topic`: Name of the topic (default: "output")
- `-b, --brokerList`: Kafka broker list (default: "localhost:9092")
- `--schemaId`: Schema ID
- `--schemaRegistryURL`: Schema Registry URL (default: "http://localhost:8081")
- `-m, --msg`: Message
- `-f, --file`: Filename
- `-a, --auth`: Auth type (default: "wo")
- `-x, --certDir`: Directory with TLS Certificates (default: "./")

### Consumer

To run the consumer, use the following command:
```sh
./kafka-avro consumer [flags]
```
Flags:

- `-t, --topic`: Name of the topic (default: "output")
- `-b, --brokerList`: Kafka broker list (default: "localhost:9092")
- `-g, --group`: Consumer group (default: "kafka-console-avro-tools")
- `-a, --auth`: Auth type (default: "wo")
- `-x, --certDir`: Directory with TLS Certificates (default: "./")
- `--schemaRegistryURL`: Schema Registry URL (default: "http://localhost:8081")

### License

This project is licensed under the MIT License.
