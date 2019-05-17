package main

import (
	"database/sql"
	"time"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"flag"
	"fmt"
	meter "github.com/larsp/co2monitor/meter"
)

func readCO(device string) (float64, int) {
	meter := new(meter.Meter)
	err := meter.Open(device)
	if err != nil {
		panic(err.Error())
	}
	//require.NoError(t, err)
	defer meter.Close()
	result, err := meter.Read()
	//require.NoError(t, err)
	log.Printf("Temp: '%v', CO2: '%v'", result.Temperature, result.Co2)
	return result.Temperature, result.Co2
}

func main() {
	dbHostPtr := flag.String("h", "localhost", "DB host")
	dbUserPtr := flag.String("u", "user", "DB user")
	dbPasswordPtr := flag.String("p", "password", "DB password")
	dbNamePtr := flag.String("n", "grafana", "DB name")
	devicePtr := flag.String("d", "/dev/hidraw0", "Device name")
	sleepTimePtr := flag.Int("sleep", 10, "Sleep time")
	flag.Parse()

	connectionString := fmt.Sprintf("%s:%s@tcp(%s:3306)/%s", *dbUserPtr, *dbPasswordPtr, *dbHostPtr, *dbNamePtr)
	log.Printf(connectionString)
	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()
	// Prepare statement for inserting data
	stmtIns, err := db.Prepare("INSERT INTO co2_state (time, temperature, co2) VALUES (?, ?, ?)")
	if err != nil {
		panic(err.Error()) 
	}
	// Connect to CO2 device
	meter := new(meter.Meter)
	meter_err := meter.Open(*devicePtr)
	if meter_err != nil {
		panic(meter_err.Error())
	}


	for {
		dt := time.Now()
		result, err := meter.Read()
		log.Printf("Temp: '%v', CO2: '%v'", result.Temperature, result.Co2)
		_, err = stmtIns.Exec(dt.UTC().Format("2006-01-02 15:04:05"), result.Temperature, result.Co2)
		if err != nil {
			panic(err.Error())
		}
	time.Sleep(time.Duration(*sleepTimePtr) * 1000 * time.Millisecond)
	}
	defer stmtIns.Close()
	defer meter.Close()
}
