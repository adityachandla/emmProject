package main

import (
	"container/heap"
	"github.com/adityachandla/emmTrial/reader"
	"github.com/adityachandla/emmTrial/search"
	"log"
)

func main() {
	houses := reader.ReadHouses()

	searchRes := search.BfsEvaluate(houses)
	for searchRes.Len() > 0 {
		node := heap.Pop(searchRes).(*search.Node)
		log.Printf("%f %s %d", node.Score, node.Conditions, node.Size)
	}
}
