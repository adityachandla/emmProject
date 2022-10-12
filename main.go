package main

import (
	"container/heap"
	"fmt"
	"github.com/adityachandla/emmTrial/reader"
	"github.com/adityachandla/emmTrial/search"
	"math"
)

func main() {
	houses := reader.ReadHouses()

	searchRes := search.BfsEvaluate(houses)
	fmt.Printf("Score   ScoreComplement   Difference   Conditions      Subgroup Size\n")
	for searchRes.Len() > 0 {
		node := heap.Pop(searchRes).(*search.Node)
		fmt.Printf("%7f %7f %7f %80s %5d\n", node.Score, node.ScoreComplement, math.Abs(node.Score-node.ScoreComplement), node.Conditions, node.Size)
	}
}
