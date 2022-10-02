package main

import (
	"container/heap"
	"fmt"
	"github.com/adityachandla/emmTrial/reader"
	"github.com/adityachandla/emmTrial/search"
)

func main() {
	houses := reader.ReadHouses()

	scoreHeap, subgroupHeap := search.BfsEvaluate(houses)
	fmt.Printf("%7s %80s %5s\n", "Score", "Conditions", "Size")
	for scoreHeap.Len() > 0 {
		node := heap.Pop(scoreHeap).(*search.Node)
		fmt.Printf("%7f %80s %5d\n", node.Score, node.Conditions, node.Size)
	}
	fmt.Printf("\n\n%7s %80s %20s %5s\n", "Score", "Conditions", "Exceptional within", "Size")
	for subgroupHeap.Len() > 0 {
		node := heap.Pop(subgroupHeap).(*search.Node)
		fmt.Printf("%7f %80s %20s %5d\n", node.SubgroupScore, node.Conditions, node.SubgroupWithin, node.Size)
	}
}
