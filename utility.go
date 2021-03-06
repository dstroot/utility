// The MIT License (MIT)
//
// Copyright (c) 2017 Daniel J. Stroot
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

// Package utility is a small library of go utility functions.  It can round
// floats, durations, check equivalency of floats, generate random hexadecimal
// strings, find the index value of a given slice value, make file names, etc.
package utility

import (
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"time"
)

const (
	timeFormat = "2006-01-02T15-04-05.000"
	dateFormat = "060102"
)

// Check streamlines error checks - use it *only*
// when the program should halt
func Check(e error) {
	if e != nil {
		log.Printf("FATAL: %+v\n", e)
		os.Exit(1)
	}
}

// SliceIndex figures out the index of slice given a value in the slice.
// http://stackoverflow.com/questions/8307478/go-how-to-find-out-element-position-in-slice
// It's used to find the slice position of the process activity step.
// Then we use that to calculate the next action.  Usage:
//     xs := []int{2, 4, 6, 8}
//     ys := []string{"C", "B", "K", "A"}
//     fmt.Println(
//         SliceIndex(len(xs), func(i int) bool { return xs[i] == 5 }),
//         SliceIndex(len(xs), func(i int) bool { return xs[i] == 6 }),
//         SliceIndex(len(ys), func(i int) bool { return ys[i] == "Z" }),
//         SliceIndex(len(ys), func(i int) bool { return ys[i] == "A" })
//     )
func SliceIndex(limit int, predicate func(i int) bool) (int, error) {
	for i := 0; i < limit; i++ {
		if predicate(i) {
			return i, nil
		}
	}
	return -1, errors.New("not found in slice")
}

// RoundFloat64 rounds numbers
func RoundFloat64(val float64, places int) float64 {
	var round float64
	pow := math.Pow(10, float64(places))
	digit := pow * val
	_, div := math.Modf(digit)
	if div >= float64(.5) {
		round = math.Ceil(digit)
	} else {
		round = math.Floor(digit)
	}
	return round / pow
}

// RoundDuration rounds time.Duration
// Example to seconds: RoundDuration(d, time.Second)
func RoundDuration(d, r time.Duration) time.Duration {
	if r <= 0 {
		return d
	}
	neg := d < 0
	if neg {
		d = -d
	}
	if m := d % r; m+m < r {
		d = d - m
	} else {
		d = d + r - m
	}
	if neg {
		return -d
	}
	return d
}

// Float64Equal checks not whether the numbers are exactly the same, but
// whether their difference is very small. The error margin that the
// difference is compared to is called epsilon.
func Float64Equal(a, b float64) bool {
	const epsilon float64 = 0.00000001
	if (a-b) < epsilon && (b-a) < epsilon {
		return true
	}
	return false
}

// MakeFileName creates a new filename
func MakeFileName(fileExtension string, directory string) string {
	t := time.Now().UTC()
	timestamp := t.Format(timeFormat)
	return filepath.Join(directory, fmt.Sprintf("%s%s", timestamp, fileExtension))
}

// GetLocalIP returns the non-loopback local IP of the host
func GetLocalIP() (ip string, err error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}
	for _, address := range addrs {
		// check the address type (not a loopback)
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ip = ipnet.IP.String()
				return ip, nil
			}
		}
	}
	return "", errors.New("no ip address found")
}

// Padding takes a string, length and either "left" or "right" for
// padding/justification plus a padding character.  It will pad the string out.
func Padding(s string, length int, justified string, padChar string) (string, error) {
	paddingNeeded := length - len(s)
	paddedString := ""

	// Check for errors
	if len(s) > length {
		return "", errors.New("string is too long")
	}
	if len(padChar) > 1 {
		return "", errors.New("padding must be only one character")
	}

	if justified == "right" {
		for i := 0; i < paddingNeeded; i++ {
			paddedString = paddedString + padChar
		}
		paddedString = paddedString + s
		return paddedString, nil
	}

	if justified == "left" {
		paddedString = s
		for i := 0; i < paddingNeeded; i++ {
			paddedString = paddedString + padChar
		}
		return paddedString, nil
	}

	return "", errors.New("justification must be either right or left")
}

// CalcSettlementDate takes in today's date (basically "now") along with a
// a map of bank holidays and it calculates to correct settlement date for
// ACH transactions.
func CalcSettlementDate(today time.Time, bankHolidayMap map[string]bool) time.Time {

	// settlement is tomorrow.
	settlementDate := today.AddDate(0, 0, 1)

	// unless tomorrow is Saturday, then it's Monday
	if "Saturday" == settlementDate.Weekday().String() {
		// add two more days (cover the weekend)
		settlementDate = settlementDate.AddDate(0, 0, 2)
	}

	// unless tomorrow is Sunday, then it's Monday
	if "Sunday" == settlementDate.Weekday().String() {
		// add one more day (cover the weekend)
		settlementDate = settlementDate.AddDate(0, 0, 1)
	}

	// unless the calculated settlement day is a
	// holiday, then add one more day.
	date := settlementDate.Format(dateFormat)
	_, found := bankHolidayMap[date]
	if found {
		settlementDate = settlementDate.AddDate(0, 0, 1)

		if "Saturday" == settlementDate.Weekday().String() {
			// add two more days (cover the weekend)
			settlementDate = settlementDate.AddDate(0, 0, 2)
		}

		if "Sunday" == settlementDate.Weekday().String() {
			// add one more day (cover the weekend)
			settlementDate = settlementDate.AddDate(0, 0, 1)
		}
	}
	return settlementDate
}

// GenRandomString returns a hexadecimal random string
func GenRandomString(strlen int) string {
	if strlen <= 0 {
		return ""
	}

	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	rand.Seed(time.Now().UTC().UnixNano())
	result := make([]byte, strlen)

	// make randon string
	for i := 0; i < strlen; i++ {
		result[i] = chars[rand.Intn(len(chars))]
	}

	// hex encode string
	hexString := hex.EncodeToString(result)
	return hexString
}
