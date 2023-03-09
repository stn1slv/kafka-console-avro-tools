# kafka-console-avro-tools

This is a CLI tool to send and receive Avro messages to and from Apache Kafka.

### Build executable file

```
go build -o kafka-console-avro-tools
```

### Producer

```
./kafka-console-avro-tools --mode=producer --brokerList=localhost:9092 --topic=topicName --schemaRegistryURL=http://localhost:8081 --schemaId=10 --msg='{"name":"Stas","address":"Moscow"}'

```

### Consumer
```
./kafka-console-avro-tools  --mode=consumer --brokerList=localhost:9092 --topic=topicName --schemaRegistryURL=http://localhost:8081
```
