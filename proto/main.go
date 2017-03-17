package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"os"
)

type Header struct {
	MagicString    string
	Version        byte
	NumRecords     uint32
}

const (
	DEBIT          byte = 0
	CREDIT         byte = 1
	START_AUTOPLAY byte = 2
	END_AUTOPLAY   byte = 3
)

type Record struct {
	RecordType     byte
	TimeStamp      uint32
	UserID         uint64
	DollarAmt      float64
}

func readBytes(file *os.File, number int) []byte {
	bytes := make([]byte, number)

	_, err := file.Read(bytes)
	if err != nil {
		log.Fatal(err)
	}

	return bytes
}

func loadHeader(file *os.File) *Header {
	header := new(Header)

	header.MagicString = string(readBytes(file, 4))
	if header.MagicString != "MPS7" {
		log.Fatal("bad file header")
	}

	header.Version = readBytes(file, 1)[0]
	numRecordsBytes := readBytes(file, 4)
	header.NumRecords = binary.BigEndian.Uint32(numRecordsBytes) //network byte order = big endian

	return header
}

func loadRecord(file *os.File) *Record {
	record := new(Record)

	record.RecordType = readBytes(file, 1)[0]
	if record.RecordType > 4 || record.RecordType < 0 {
		log.Fatal("invalid record type")
	}

	timeStampBytes := readBytes(file, 4)
	record.TimeStamp = binary.BigEndian.Uint32(timeStampBytes)

	userIDBytes := readBytes(file, 8)
	record.UserID = binary.BigEndian.Uint64(userIDBytes)

	if record.RecordType == DEBIT || record.RecordType == CREDIT {
		dollarAmtBytes := readBytes(file, 8)
		record.DollarAmt = math.Float64frombits(binary.BigEndian.Uint64(dollarAmtBytes))
	} else {
		record.DollarAmt = 0.0
	}

	return record
}

func main() {
	path := "txnlog.dat"
	totalDebit := float64(0)
	totalCredit := float64(0)
	totalAutopayStart := 0
	totalAutopayEnd := 0
	user_balance := float64(0)

	file, err := os.Open(path)
	if err != nil {
		log.Fatal("Error opening file ", err)
	}

	header := loadHeader(file)

	for i := uint32(0); i < header.NumRecords; i++ {
		record := loadRecord(file)

		if record.UserID == 2456938384156277127 {
			user_balance += record.DollarAmt
		}

		switch record.RecordType {
		case DEBIT:
			totalDebit += record.DollarAmt
		case CREDIT:
			totalCredit += record.DollarAmt
		case START_AUTOPLAY:
			totalAutopayStart++
		case END_AUTOPLAY:
			totalAutopayEnd++
		}
	}

	fmt.Printf("What is the total amount in dollars of debits?      %f\n", totalDebit)
	fmt.Printf("What is the total amount in dollars of credits?     %f\n", totalCredit)
	fmt.Printf("How many autopays were started?                     %d\n", totalAutopayStart)
	fmt.Printf("How many autopays were ended?                       %d\n", totalAutopayEnd)
	fmt.Printf("What is balance of user ID 2456938384156277127?     %f\n", user_balance)

	defer file.Close()
}
