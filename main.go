package main

import (
	"github.com/adityachandla/emmTrial/reader"
	"github.com/adityachandla/emmTrial/search"
	"log"
)

func main() {
	houses := reader.ReadHouses()

	searchRes := search.BfsEvaluate(houses)
	for _, res := range searchRes {
		log.Printf("%s %f\n", res.Conditions, res.Score)
	}
	score, _ := search.CalculateCorrelation(houses, nil)
	log.Printf("Comparision to: %f", score)
}
