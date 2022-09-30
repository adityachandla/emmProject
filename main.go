package main

import (
	"log"
)

const (
	Depth      int     = 3
	MinSupport float64 = 0.1
)

var houses []*HouseInfo

var stringFields = []string{"HouseType", "Suburb"}
var intFields = []string{"Car", "Rooms"}

//var intFields = []string{"Bedrooms", "Bathrooms", "Car", "Rooms"}

var stringTargets []*stringTarget
var intTargets []*intTarget

var searchRes = make(map[string]*searchNode)

func main() {
	houses = ReadHouses()
	stringTargets = calculateStringTargets()
	intTargets = calculateIntTargets()

	bfsEvaluate()
	for _, res := range searchRes {
		log.Printf("%s %f\n", res.conditions, res.score)
	}
	log.Printf("Comparision to: %f", calculateCorrelation(nil))
}
