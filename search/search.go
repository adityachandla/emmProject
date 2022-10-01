// Package search
// This package is used for doing a breadth first search on the conditions in our dataset.
// In this file, we define the strategy for doing the breadth first search.
package search

import (
	"github.com/adityachandla/emmTrial/reader"
	"github.com/gammazero/deque"
)

const (
	Depth      int     = 3
	MinSupport float64 = 0.1
)

var (
	stringTargets []*StringTarget
	intTargets    []*IntTarget

	searchRes map[string]*Node
	houses    []*reader.HouseInfo
)

func BfsEvaluate(h []*reader.HouseInfo) map[string]*Node {
	houses = h

	stringTargets = calculateStringTargets(houses)
	intTargets = calculateIntTargets(houses)

	searchRes = make(map[string]*Node)

	queue := deque.New[*Node](16)
	queue.PushBack(&Node{})
	for queue.Len() != 0 {
		curr := queue.PopFront()
		//The number of Conditions denotes the depth.
		//If we have more Conditions than the depth then we can stop
		if len(curr.Conditions) == Depth {
			continue
		}
		processStringTargets(curr, queue)
		//We process strings in both increasing and decreasing order
		//So if we have 1,2,3 for a field then less than equal
		//conditions would be <=1, <=2, <=3
		processIntTargetsLessThanEqual(curr, queue)
		//and greater than equal conditions would be >=1, >=2,>=3
		processIntTargetsGreaterThanEqual(curr, queue)
	}
	return searchRes
}

func processStringTargets(curr *Node, queue *deque.Deque[*Node]) {
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
			if !hasSupport(size, len(houses)) {
				continue
			}
			searchRes[nextNode.Conditions.String()] = nextNode
			queue.PushBack(nextNode)
		}
	}
}

func processIntTargetsLessThanEqual(curr *Node, queue *deque.Deque[*Node]) {
	for targetIdx := curr.intTargetStartIdx; targetIdx < len(intTargets); targetIdx++ {
		targetToAdd := intTargets[targetIdx]
		//go from start to end accumulating count and only branch out when
		//count is greater than minimum support
		frequency := 0
		for i := targetToAdd.MinVal + 1; i <= targetToAdd.MaxVal; i += 2 {
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
			if !hasSupport(size, len(houses)) {
				continue
			}
			queue.PushBack(nextNode)
			searchRes[nextNode.Conditions.String()] = nextNode
		}
	}
}

func processIntTargetsGreaterThanEqual(curr *Node, queue *deque.Deque[*Node]) {
	for targetIdx := curr.intTargetStartIdx; targetIdx < len(intTargets); targetIdx++ {
		targetToAdd := intTargets[targetIdx]
		//go from end to start and only branch out when
		//count is greater than minimum support
		frequency := 0
		for i := targetToAdd.MaxVal - 1; i >= targetToAdd.MinVal; i-- {
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
			if !hasSupport(size, len(houses)) {
				continue
			}
			queue.PushBack(nextNode)
			searchRes[nextNode.Conditions.String()] = nextNode
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
