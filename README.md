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
    - For Confluent Schema Registry, please specify "http://MY-REGISTRY-URL"
    - For Apicurio Registry, please specify "http://MY-REGISTRY-URL/apis/ccompat/v7"
- `-m, --msg`: Message
- `-f, --file`: Filename
- `-a, --auth`: Auth type (default: "wo"). Supported options:
    - `wo` - without authentication
    - `tls` - TLS/SSL authentication
- `--certFile`: TLS certificate file (in pem format) (default: "./client.cer.pem")
- `--keyFile`: TLS key file (in pem format) (default: "./client.key.pem")
- `--caCertFile`: TLS CA certificate file (in pem format) (default: "./server.cer.pem")

### Consumer

To run the consumer, use the following command:
```sh
./kafka-avro consumer [flags]
```
Flags:

- `-t, --topic`: Name of the topic (default: "output")
- `-b, --brokerList`: Kafka broker list (default: "localhost:9092")
- `-g, --group`: Consumer group (default: "kafka-console-avro-tools")
- `--schemaRegistryURL`: Schema Registry URL (default: "http://localhost:8081")
    - For Confluent Schema Registry, please specify "http://MY-REGISTRY-URL"
    - For Apicurio Registry, please specify "http://MY-REGISTRY-URL/apis/ccompat/v7"
- `-a, --auth`: Auth type (default: "wo"). Supported options:
    - `wo` - without authentication
    - `tls` - TLS/SSL authentication 
- `--certFile`: TLS certificate file (in pem format) (default: "./client.cer.pem")
- `--keyFile`: TLS key file (in pem format) (default: "./client.key.pem")
- `--caCertFile`: TLS CA certificate file (in pem format) (default: "./server.cer.pem")

### Testing

1. Register simple Avro schema
```sh
curl -s -L 'http://localhost:8081/subjects/Kafka-value/versions' \
-H 'Content-Type: application/vnd.schemaregistry.v1+json' \
-d '{"schema": "{\"type\":\"record\",\"name\":\"Record\",\"fields\":[{\"name\":\"name\",\"type\":\"string\"},{\"name\":\"address\",\"type\":\"string\"}]}"}'
```
2. Produce a message to kafka topic
```sh
./kafka-avro producer --schemaId=1 --file examples/example.json
```
3. Consume messages from kafka topic
```sh
./kafka-avro consumer
```

### License

This project is licensed under the MIT License.
