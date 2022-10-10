// Package search
// This package is used for doing a breadth first search on the conditions in our dataset.
// In this file, we define the strategy for doing the breadth first search.
package search

import (
	"container/heap"
	"github.com/adityachandla/emmTrial/reader"
	"github.com/gammazero/deque"
	"log"
	"math"
	"sync"
)

const (
	Depth      int     = 4
	MinSupport float64 = 0.01
	MaxLen     int     = 10
	NumWorkers int     = 8
)

var (
	houses []*reader.HouseInfo

	stringTargets []*StringTarget
	intTargets    []*IntTarget

	baseScore float64
)

func initializeGlobalVariables(h []*reader.HouseInfo) {
	houses = h
	baseScore, _ = CalculateCorrelation(houses, nil)
	log.Printf("Base score is %f", baseScore)

	stringTargets = calculateStringTargets(houses)
	intTargets = calculateIntTargets(houses)
}

func BfsEvaluate(h []*reader.HouseInfo) *NodeHeap {
	initializeGlobalVariables(h)
	queue := deque.Deque[*Node]{}
	nodeHeap := &NodeHeap{}
	queue.PushBack(&Node{})
	for depth := 0; depth <= Depth && queue.Len() > 0; depth++ {
		log.Println("Processing level ", depth)
		outputChannel := make(chan *Node, 5_000)
		inputChannel := make(chan *Node, NumWorkers)
		wg := sync.WaitGroup{}
		wg.Add(NumWorkers)
		for i := 0; i < NumWorkers; i++ {
			go func() {
				defer wg.Done()
				worker(inputChannel, outputChannel)
			}()
		}
		go processLevel(outputChannel, &queue, nodeHeap)
		for queue.Len() > 0 {
			node := queue.PopBack()
			inputChannel <- node
		}
		close(inputChannel)
		wg.Wait()
		close(outputChannel)
	}

	return nodeHeap
}

func worker(input <-chan *Node, output chan<- *Node) {
	for curr := range input {
		if curr != nil {
			processStringTargets(curr, output)
			processIntTargetsLessThanEqual(curr, output)
			processIntTargetsGreaterThanEqual(curr, output)
		}
	}
}

func processLevel(outputChannel <-chan *Node, queue *deque.Deque[*Node], nodeHeap *NodeHeap) {
	seenNodes := make(map[string]struct{})
	for node := range outputChannel {
		if _, present := seenNodes[node.Conditions.String()]; present {
			continue
		}
		queue.PushBack(node)
		if nodeHeap.Len() < MaxLen {
			heap.Push(nodeHeap, node)
		} else if node.Score > (*nodeHeap)[0].Score {
			heap.Pop(nodeHeap)
			heap.Push(nodeHeap, node)
		}
		seenNodes[node.Conditions.String()] = struct{}{}
	}
}

func processStringTargets(curr *Node, outputChannel chan<- *Node) {
	for targetIdx := curr.stringTargetStartIdx; targetIdx < len(stringTargets); targetIdx++ {
		targetToAdd := stringTargets[targetIdx]
		for targetValue, count := range targetToAdd.Values {
			//This is just a preemptive pruning because we don't consider the fact that curr will already
			//be a subset of all the houses. The actual pruning happens after we have the score and count.
			if !hasSupport(count, len(houses)) {
				continue
			}
			newCondition := getStringEqualityCondition(targetToAdd.FieldName, targetValue)
			nextNode := getNodeWithAddedCondition(newCondition, curr)
			nextNode.stringTargetStartIdx = targetIdx + 1
			processNode(nextNode, outputChannel)
		}
	}
}

func getStringEqualityCondition(name, value string) *Condition {
	return &Condition{
		IsString:         true,
		FieldName:        name,
		FieldValueString: value,
		Inequality:       Equal,
	}
}

func processIntTargetsLessThanEqual(curr *Node, outputChannel chan<- *Node) {
	for targetIdx := curr.intTargetStartIdx; targetIdx < len(intTargets); targetIdx++ {
		targetToAdd := intTargets[targetIdx]
		//go from start to end accumulating count and only branch out when
		//count is greater than minimum support
		frequency := 0
		for i := targetToAdd.MinVal + 1; i <= targetToAdd.MaxVal-1; i += 2 {
			frequency += targetToAdd.Counter[i] + targetToAdd.Counter[i-1]
			if !hasSupport(frequency, len(houses)) {
				continue
			}
			newCondition := getIntLessThanEqualCondition(targetToAdd.FieldName, i)
			nextNode := getNodeWithAddedCondition(newCondition, curr)
			nextNode.intTargetStartIdx = targetIdx + 1
			processNode(nextNode, outputChannel)
		}
	}
}

func getIntLessThanEqualCondition(name string, value int) *Condition {
	return &Condition{
		FieldName:     name,
		FieldValueInt: value,
		Inequality:    LessThanEqual,
		IsString:      false,
	}
}

func processIntTargetsGreaterThanEqual(curr *Node, outputChannel chan<- *Node) {
	for targetIdx := curr.intTargetStartIdx; targetIdx < len(intTargets); targetIdx++ {
		targetToAdd := intTargets[targetIdx]
		//go from end to start and only branch out when
		//count is greater than minimum support
		frequency := 0
		for i := targetToAdd.MaxVal - 1; i >= targetToAdd.MinVal+1; i-- {
			frequency += targetToAdd.Counter[i] + targetToAdd.Counter[i+1]
			if !hasSupport(frequency, len(houses)) {
				continue
			}
			newCondition := getIntGreaterThanEqualCondition(targetToAdd.FieldName, i)
			nextNode := getNodeWithAddedCondition(newCondition, curr)
			nextNode.intTargetStartIdx = targetIdx + 1
			processNode(nextNode, outputChannel)
		}
	}
}

func getIntGreaterThanEqualCondition(name string, value int) *Condition {
	return &Condition{
		FieldName:     name,
		FieldValueInt: value,
		Inequality:    GreaterThanEqual,
		IsString:      false,
	}
}

func processNode(nextNode *Node, outputChannel chan<- *Node) {
	nextNode.Correlation, nextNode.Size = CalculateCorrelation(houses, nextNode.Conditions)
	nextNode.Score = math.Abs(nextNode.Correlation - baseScore)
	if hasSupport(nextNode.Size, len(houses)) {
		outputChannel <- nextNode
	}
}

func getNodeWithAddedCondition(condition *Condition, curr *Node) *Node {
	newConditions := copySlice(curr.Conditions)
	newConditions = append(newConditions, condition)
	return &Node{
		Conditions:           newConditions,
		stringTargetStartIdx: curr.stringTargetStartIdx,
		intTargetStartIdx:    curr.intTargetStartIdx,
	}
}

func copySlice[T any](slice []T) []T {
	if slice == nil {
		return make([]T, 0, Depth)
	}
	newSlice := make([]T, len(slice))
	copy(newSlice, slice)
	return newSlice
}

func hasSupport(count, total int) bool {
	return (float64(count) / float64(total)) > MinSupport
}
