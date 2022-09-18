package main

import "testing"

func BenchmarkReading(b *testing.B) {
	readHouses()
}
