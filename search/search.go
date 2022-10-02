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
	MaxLen     int     = 40
)

// TODO Create a struct for all this
var (
	houses []*reader.HouseInfo

	stringTargets []*StringTarget
	intTargets    []*IntTarget

	nodeHeap  *NodeHeap
	baseScore float64
	seenNodes map[string]struct{}
	queue     *deque.Deque[*Node]
	mutex     sync.Mutex
)

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
}

func BfsEvaluate(h []*reader.HouseInfo) *NodeHeap {
	houses = h
	baseScore, _ = CalculateCorrelation(houses, nil)
	log.Printf("Base score is %f", baseScore)

	stringTargets = calculateStringTargets(houses)
	intTargets = calculateIntTargets(houses)

	nodeHeap = &NodeHeap{}
	seenNodes = make(map[string]struct{})

	queue = deque.New[*Node](16)
	queue.PushBack(&Node{})
	for queue.Len() != 0 {
		curr := queue.PopFront()
		//The number of Conditions denotes the depth.
		//If we have more Conditions than the depth then we can stop
		if len(curr.Conditions) == Depth {
			continue
		}
		wg := &sync.WaitGroup{}
		wg.Add(3)
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
		go func() {
			defer wg.Done()
			processIntTargetsGreaterThanEqual(curr)
		}()
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
			newConditions := copySlice(curr.Conditions)
			newConditions = append(newConditions, &Condition{
				IsString:         true,
				FieldName:        targetToAdd.FieldName,
				FieldValueString: targetValue,
				Inequality:       Equal, //We process the string targets only for equality.
			})
			nextNode := &Node{
				Conditions:           newConditions,
				stringTargetStartIdx: targetIdx + 1,
				intTargetStartIdx:    curr.intTargetStartIdx,
			}
			var size int
			nextNode.Score, size = CalculateCorrelation(houses, nextNode.Conditions)
			nextNode.Score = math.Abs(nextNode.Score - baseScore)
			nextNode.Size = size
			if hasSupport(size, len(houses)) {
				addNode(nextNode)
			}
		}
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
			//less than equal
			newConditions := copySlice(curr.Conditions)
			newConditions = append(newConditions, &Condition{
				FieldName:     targetToAdd.FieldName,
				FieldValueInt: i,
				Inequality:    LessThanEqual,
				IsString:      false,
			})
			nextNode := &Node{
				intTargetStartIdx:    targetIdx + 1,
				stringTargetStartIdx: curr.stringTargetStartIdx,
				Conditions:           newConditions,
			}
			var size int
			nextNode.Score, size = CalculateCorrelation(houses, newConditions)
			nextNode.Score = math.Abs(nextNode.Score - baseScore)
			nextNode.Size = size
			if hasSupport(size, len(houses)) {
				addNode(nextNode)
			}
		}
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
			newConditions := copySlice(curr.Conditions)
			newConditions = append(newConditions, &Condition{
				FieldName:     targetToAdd.FieldName,
				FieldValueInt: i,
				Inequality:    GreaterThanEqual,
				IsString:      false,
			})
			nextNode := &Node{
				intTargetStartIdx:    targetIdx + 1,
				stringTargetStartIdx: curr.stringTargetStartIdx,
				Conditions:           newConditions,
			}
			var size int
			nextNode.Score, size = CalculateCorrelation(houses, newConditions)
			nextNode.Score = math.Abs(nextNode.Score - baseScore)
			nextNode.Size = size
			if hasSupport(size, len(houses)) {
				addNode(nextNode)
			}
		}
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
