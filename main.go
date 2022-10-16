package main

import (
	"container/heap"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/adityachandla/emmTrial/reader"
	"github.com/adityachandla/emmTrial/search"
)

func main() {
	houses := reader.ReadHouses()

	searchRes := search.BfsEvaluate(houses)

	// Initialize tabWriter for column formatting
	w := new(tabwriter.Writer)

	// minwidth, tabwidth, padding, padchar, flags
	w.Init(os.Stdout, 8, 8, 2, '\t', 0)

	defer w.Flush()

	fmt.Fprintf(w, "Score\tScoreComplement\tDifference\tScoreRelative\tConditions\tSubgroup Size\n")
	fmt.Fprintf(w, "----\t----\t----\t----\t----\t----\n")

	for searchRes.Len() > 0 {
		node := heap.Pop(searchRes).(*search.Node)
		fmt.Fprintf(w, "%f\t%f\t%f\t%f\t%s\t%d\n", node.Score, node.ScoreComplement, node.ScoreDifference, node.ScoreRelative, node.Conditions, node.Size)
	}
}
