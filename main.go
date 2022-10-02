package main

import (
	"container/heap"
	"fmt"
	"github.com/adityachandla/emmTrial/reader"
	"github.com/adityachandla/emmTrial/search"
)

func main() {
	houses := reader.ReadHouses()

	searchRes := search.BfsEvaluate(houses)
	fmt.Printf("Score   Conditions      Subgroup Size\n")
	for searchRes.Len() > 0 {
		node := heap.Pop(searchRes).(*search.Node)
		fmt.Printf("%7f %160s %5d\n", node.Score, node.Conditions, node.Size)
	}
}
