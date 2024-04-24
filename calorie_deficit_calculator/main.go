package main

import (
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	common "github.com/gelberg/calorie-deficit-tracker/common"
)

const (
	ReportIntervalParamName        = "DEFICIT_REPORT_INTERVAL_MS"
	DefaultDeficitReportIntervalMs = 10000
)

var (
	consumption = 0
	expenditure = 0
	deficit     = 0
	lock        = sync.Mutex{}
)

func readConsumption() {
	topic := "consumption"

	conn, err := common.ConnectToKafka(topic)
	if err != nil {
		log.Fatal(err)
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
}

func readExpenditure() {
	topic := "expenditure"

	conn, err := common.ConnectToKafka(topic)
	if err != nil {
		log.Fatal(err)
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
}

func main() {
	go readConsumption()
	go readExpenditure()

	reportIntervalMs := common.GetEnvInt(ReportIntervalParamName, DefaultDeficitReportIntervalMs)
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
