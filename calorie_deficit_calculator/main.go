package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/segmentio/kafka-go"
)

var (
	kafkaEndpoint           = os.Getenv("KAFKA_ENDPOINT")
	deficitReportIntervalMs = os.Getenv("DEFICIT_REPORT_INTERVAL_MS")
	consumption             = 0
	expenditure             = 0
	deficit                 = 0
	lock                    = sync.Mutex{}
)

func connectToKafka(topic string) (*kafka.Conn, error) {
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

	return conn, err
}

func readConsumption() {
	topic := "consumption"

	conn, err := connectToKafka(topic)
	if err != nil {
		log.Fatal("failed to dial leader:", err)
	}

	for {
		conn.SetReadDeadline(time.Now().Add(75 * time.Second))
		msg, err := conn.ReadMessage(10)
		if err != nil {
			log.Println(err)
		}
		cons, err := strconv.Atoi(string(msg.Value[:]))
		if err != nil {
			log.Println(err)
			continue
		}

		lock.Lock()
		consumption = cons
		deficit = expenditure - consumption
		lock.Unlock()
	}

	// if err := conn.Close(); err != nil {
	// 	log.Fatal("failed to close connection:", err)
	// }
}

func readExpenditure() {
	topic := "expenditure"

	conn, err := connectToKafka(topic)
	if err != nil {
		log.Fatal("failed to dial leader:", err)
	}

	for {
		conn.SetReadDeadline(time.Now().Add(10 * time.Second))
		msg, err := conn.ReadMessage(10)
		if err != nil {
			log.Println(err)
		}
		exp, err := strconv.Atoi(string(msg.Value[:]))
		if err != nil {
			log.Println(err)
			continue
		}

		lock.Lock()
		expenditure = exp
		deficit = expenditure - consumption
		lock.Unlock()
	}

	// if err := conn.Close(); err != nil {
	// 	log.Fatal("failed to close connection:", err)
	// }
}

func main() {
	go readConsumption()
	go readExpenditure()

	reportIntervalMs, err := strconv.Atoi(deficitReportIntervalMs)
	if err != nil {
		log.Fatal(err)
	}

	reportInterval := time.Millisecond * time.Duration(reportIntervalMs)
	nextReport := time.Now()

	for {
		lock.Lock()
		fmt.Printf("Deficit at %v: %v\n", time.Now(), deficit)
		lock.Unlock()

		nextReport = nextReport.Add(reportInterval)
		time.Sleep(-time.Since(nextReport))
	}
}
