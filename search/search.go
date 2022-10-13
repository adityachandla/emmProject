// Package search
// This package is used for doing a breadth first search on the conditions in our dataset.
// In this file, we define the strategy for doing the breadth first search.
package search

import (
	"container/heap"
	"fmt"
	"log"
	"math"
	"sync"

	"github.com/adityachandla/emmTrial/reader"
	"github.com/gammazero/deque"
)

const (
	Depth      int     = 4
	MinSupport float64 = 0.01
	MaxLen     int     = 40
	X          float64 = 0.1
)

// TODO should I create a struct for all this?
var (
	processedCounter int
	houses           []*reader.HouseInfo

	stringTargets []*StringTarget
	intTargets    []*IntTarget

	nodeHeap  *NodeHeap
	baseScore float64
	seenNodes map[string]struct{}
	queue     *deque.Deque[*Node]
	mutex     sync.Mutex
)

func initializeGlobalVariables(h []*reader.HouseInfo) {
	houses = h
	baseScore, _ = CalculateCorrelation(houses, nil)
	log.Printf("Base score is %f", baseScore)

	stringTargets = calculateStringTargets(houses)
	intTargets = calculateIntTargets(houses)

	nodeHeap = &NodeHeap{}
	seenNodes = make(map[string]struct{})

	queue = deque.New[*Node](16)
}

func BfsEvaluate(h []*reader.HouseInfo) *NodeHeap {
	initializeGlobalVariables(h)
	//Start with no conditions
	queue.PushBack(&Node{})
	for queue.Len() != 0 {
		curr := queue.PopFront()
		//The number of Conditions denotes the depth.
		//If we have more Conditions than the depth then we can stop
		if len(curr.Conditions) == Depth {
			continue
		}
		wg := &sync.WaitGroup{}
		wg.Add(2)
		go func() {
			defer wg.Done()
			processStringTargets(curr)
		}()
		//We process strings in both increasing and decreasing order
		//So if we have 1,2,3 for a field then less than equal
		//conditions would be <=1, <=2, <=3
		go func() {
			defer wg.Done()
			processIntTargetsLessThanEqual(curr)
		}()
		//and greater than equal conditions would be >=1, >=2,>=3
		processIntTargetsGreaterThanEqual(curr)
		wg.Wait()
	}
	return nodeHeap
}

func processStringTargets(curr *Node) {
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
			processNode(nextNode)
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

func processIntTargetsLessThanEqual(curr *Node) {
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
			processNode(nextNode)
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

func processIntTargetsGreaterThanEqual(curr *Node) {
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
			processNode(nextNode)
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

func processNode(nextNode *Node) {
	nextNode.Correlation, nextNode.Size = CalculateCorrelation(houses, nextNode.Conditions)
	correlationComplement, _ := CalculateCorrelationComplement(houses, nextNode.Conditions)
	nextNode.ScoreComplement = math.Abs(correlationComplement - nextNode.Correlation)
	nextNode.Score = math.Abs(nextNode.Correlation - baseScore)
	nextNode.ScoreDifference = math.Abs(nextNode.Score - nextNode.ScoreComplement)
	if float64(nextNode.Size/len(houses)) > X {
		nextNode.ScoreRelative = nextNode.ScoreComplement
	} else {
		nextNode.ScoreRelative = nextNode.Score
	}
	if hasSupport(nextNode.Size, len(houses)) {
		addNode(nextNode)
	}
}

func addNode(node *Node) {
	mutex.Lock()
	defer mutex.Unlock()

	if _, present := seenNodes[node.Conditions.String()]; present {
		return
	}
	queue.PushBack(node)
	if nodeHeap.Len() < MaxLen {
		heap.Push(nodeHeap, node)
	} else if node.Score > (*nodeHeap)[0].Score {
		heap.Pop(nodeHeap)
		heap.Push(nodeHeap, node)
	}
	seenNodes[node.Conditions.String()] = struct{}{}
	fmt.Printf("Added nodes : %d\r", processedCounter)
	processedCounter++
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
