package common

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	kafka "github.com/segmentio/kafka-go"
)

var (
	kafkaEndpoint = os.Getenv("KAFKA_ENDPOINT")
)

func ConnectToKafka(topic string) (*kafka.Conn, error) {
	partition := 0
	// Workaround for the following scenario:
	// 1. docker-compose with kafka and this service is started:
	// 1.1. kafka is started
	// 1.2. this service attempts to connect to kafka, gets connection refuse and exits
	// 1.3. kafka initializes its listeners
	connectionAttempts := 3
	conn, err := kafka.DialLeader(context.Background(), "tcp", kafkaEndpoint, topic, partition)
	for connectionAttempts > 0 && err != nil { // TODO: distinguish our case and others?
		conn, err = kafka.DialLeader(context.Background(), "tcp", kafkaEndpoint, topic, partition)
		connectionAttempts--
		time.Sleep(time.Second)
	}

	if err != nil {
		return nil, fmt.Errorf("kafka.DialLeader(): %v", err)
	}

	return conn, nil
}

func GetEnvInt(param string, defaultValue int) int {
	valueStr := os.Getenv(param)
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		log.Printf("Error on obtaining parameter %v: %v. Default "+
			"interval %v ms will be used instead.", param, err, defaultValue)

		return defaultValue
	}

	return value
}
