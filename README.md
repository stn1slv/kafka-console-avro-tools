# kafka-console-avro-tools

This is a CLI tool to send and receive Avro messages to and from Apache Kafka.

### Build executable file

```
go build -o kafka-console-avro-tools
```

### Producer

##### Register demo avro schema to Confluent Schema Registry
```
curl -s -L 'http://localhost:8081/subjects/Kafka-value/versions' \
-H 'Content-Type: application/vnd.schemaregistry.v1+json' \
-d '{"schema": "{\"type\":\"record\",\"namespace\":\"example\",\"name\":\"DemoEntity\",\"fields\":[{\"name\":\"name\",\"type\":\"string\"},{\"name\":\"address\",\"type\":\"string\"}]}"}'
```
##### Send a message as a param
```
./kafka-console-avro-tools --mode=producer --topic=topicName --brokerList=localhost:9092 --schemaRegistryURL=http://localhost:8081 --schemaId=1 --msg='{"name":"Stas","address":"Moscow"}'

```

##### Send a message from file
```
./kafka-console-avro-tools --mode=producer --topic=topicName --brokerList=localhost:9092 --schemaRegistryURL=http://localhost:8081 --schemaId=1 --file=examples/file.json
```

### Consumer
```
./kafka-console-avro-tools --mode=consumer --topic=topicName --brokerList=localhost:9092 --schemaRegistryURL=http://localhost:8081
```
