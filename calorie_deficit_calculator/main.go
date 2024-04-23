package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	common "github.com/gelberg/calorie-deficit-tracker/common"
)

var (
	deficitReportIntervalMs = os.Getenv("DEFICIT_REPORT_INTERVAL_MS")
	consumption             = 0
	expenditure             = 0
	deficit                 = 0
	lock                    = sync.Mutex{}
)

func readConsumption() {
	topic := "consumption"

	conn, err := common.ConnectToKafka(topic)
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

	conn, err := common.ConnectToKafka(topic)
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
