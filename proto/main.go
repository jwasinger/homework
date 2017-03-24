package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"os"
)

type Header struct {
	MagicString string
	Version     byte
	NumRecords  uint32
}

const (
	DEBIT          byte = 0
	CREDIT         byte = 1
	START_AUTOPLAY byte = 2
	END_AUTOPLAY   byte = 3
)

/*
	MPS7 stores dollar amounts using float64 format.  However, this is not a good
	representation for precise calculation due to accumulation of floating point errors
	over successive operations on numbers.

	I chose to represent dollar amounts as int64 types.

	For example:
	the floating point value float64(23.15) would be represented as int64(2315)
*/
type Record struct {
	RecordType byte
	TimeStamp  uint32
	UserID     uint64
	DollarAmt  int64
}

/*
convert a float64 dollar amount to an int64 representation
*/
func truncate(num float64) int64 {
	return int64(num * 100.0)
}

/*
read a number of bytes from file, advance the offset by that amount
*/
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
		dollarAmt := math.Float64frombits(binary.BigEndian.Uint64(dollarAmtBytes))
		record.DollarAmt = truncate(dollarAmt)
	} else {
		record.DollarAmt = 0
	}

	return record
}

/*
	convert int64 dollar representation to printable string
*/
func formatAmt(val int64) string {
	cents := val % 100
	dollars := val / 100
	return fmt.Sprintf("%d.%d", dollars, cents)
}

func main() {
	path := "txnlog.dat"
	var totalDebit int64
	var totalCredit int64
	var totalAutopayStart int64
	var totalAutopayEnd int64
	var userBalance int64

	file, err := os.Open(path)
	if err != nil {
		log.Fatal("Error opening file ", err)
	}

	header := loadHeader(file)

	for i := uint32(0); i < header.NumRecords; i++ {
		record := loadRecord(file)

		if record.UserID == 2456938384156277127 {
			userBalance += record.DollarAmt
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

	fmt.Printf("Total amount in dollars of debits:      %s\n", formatAmt(totalDebit))
	fmt.Printf("Total amount in dollars of credits:     %s\n", formatAmt(totalCredit))
	fmt.Printf("Balance of user ID 2456938384156277127: %s\n", formatAmt(userBalance))
	fmt.Printf("Number of autopays were started:        %d\n", totalAutopayStart)
	fmt.Printf("Number of autopays were ended:          %d\n", totalAutopayEnd)

	defer file.Close()
}
