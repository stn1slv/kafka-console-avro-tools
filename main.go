package main

import (
	"log"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "kafka-avro",
		Short: "Kafka Avro producer and consumer",
	}

	producerCmd := newProducerCommand()
	consumerCmd := newConsumerCommand()

	rootCmd.AddCommand(producerCmd, consumerCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}
