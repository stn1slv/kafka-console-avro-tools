# kafka-console-avro-tools

This is a CLI tool to send and receive Avro messages to and from Apache Kafka.

### Build executable file

```
go build -o kafka-console-avro-tools
```

### Producer

```
./kafka-console-avro-tools --brokerList=localhost:9092 --schemaRegistryURL=http://localhost:8081 --mode=producer --schemaId=10 --msg='{"name":"Stas","address":"Moscow"}'

```

### Consumer
```
./kafka-console-avro-tools --brokerList=localhost:9092 --schemaRegistryURL=http://localhost:8081 --mode=consumer 
```