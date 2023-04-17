# Go Kafka Avro Tools

This is a simple Golang command-line tool to produce and consume Apache Kafka messages using the Avro serialization format. The tool interacts with the Confluent Schema Registry to store and retrieve Avro schemas.

## Prerequisites

- Go 1.16 or higher
- Running Apache Kafka cluster
- Running Confluent Schema Registry

## Installation

1. Clone this repository:

```sh
git clone https://github.com/yourusername/go-kafka-avro-tools.git
```

2. Change to the project directory:
```sh
cd go-kafka-avro-tools
```

3. Build the project:
```sh
go build -o go-kafka-avro-tools
```

4. Make sure to set up the Kafka cluster and Confluent Schema Registry.

## Usage
You can use this tool in two modes: producer and consumer.

### Producer
To produce messages, run the following command:
```sh
./go-kafka-avro-tools --mode=producer --topic=your_topic --brokerList=your_broker_list --schemaId=your_schema_id --msg="your_message"
```

- Replace `your_topic` with the target Kafka topic.
- Replace `your_broker_list` with the list of Kafka brokers (e.g., `localhost:9092`).
- Replace `your_schema_id` with the ID of the Avro schema from the Schema Registry.
- Replace `your_message` with the message you want to send.

### Consumer

To consume messages, run the following command:
```sh
./go-kafka-avro-tools --mode=consumer --topic=your_topic --brokerList=your_broker_list --group=your_consumer_group
```


- Replace `your_topic` with the target Kafka topic.
- Replace `your_broker_list` with the list of Kafka brokers (e.g., `localhost:9092`).
- Replace `your_consumer_group` with the consumer group you want to use.

## Configuration

You can configure the tool using command-line flags or environment variables. The available options are:

- `--topic` or `TOPIC`: The Kafka topic name. Default is `output`.
- `--brokerList` or `BROKER_LIST`: The list of Kafka brokers. Default is `localhost:9092`.
- `--mode` or `MODE`: The mode of the tool: `producer` or `consumer`. Default is `producer`.
- `--schemaId` or `SCHEMA_ID`: The Avro schema ID from the Schema Registry.
- `--schemaRegistryURL` or `SCHEMA_REGISTRY_URL`: The Schema Registry URL. Default is `http://localhost:8081`.
- `--msg` or `MSG`: The message to be sent as a producer.
- `--file` or `FILE`: A file containing the message to be sent as a producer.
- `--group` or `GROUP`: The consumer group for the consumer mode. Default is `kafka-console-avro-tools`.
- `--auth` or `AUTH`: The authentication type. Default is `wo`. Set to `TLS` for TLS authentication.
- `--certDir` or `TLS_CERTS_DIR`: The directory with TLS certificates. Default is `./`.