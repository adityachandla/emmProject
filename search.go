package main

import (
	"fmt"
	"github.com/gammazero/deque"
	"sort"
	"strings"
)

type inequality int

const (
	Equal inequality = iota
	LessThanEqual
	GreaterThanEqual
)

func (i inequality) String() string {
	switch i {
	case Equal:
		return "="
	case LessThanEqual:
		return "<="
	case GreaterThanEqual:
		return ">="
	}
	panic("Unexpected value for inequality")
}

type searchNode struct {
	conditions           searchConditions
	score                float64
	stringTargetStartIdx int
	intTargetStartIdx    int
}

type searchCondition struct {
	fieldName        string
	isString         bool //it will either be string or int
	fieldValueString string
	fieldValueInt    int
	inequality       inequality
}

func (cond *searchCondition) String() string {
	if cond.isString {
		return fmt.Sprintf("(%s %s %s)", cond.fieldName, cond.inequality, cond.fieldValueString)
	}
	return fmt.Sprintf("(%s %s %d)", cond.fieldName, cond.inequality, cond.fieldValueInt)
}

type searchConditions []*searchCondition

func (conditions searchConditions) Len() int { return len(conditions) }
func (conditions searchConditions) Less(i, j int) bool {
	return conditions[i].fieldName < conditions[j].fieldName
}
func (conditions searchConditions) Swap(i, j int) {
	conditions[i], conditions[j] = conditions[j], conditions[i]
}

func (conditions searchConditions) String() string {
	sort.Sort(conditions)
	sb := strings.Builder{}
	for _, field := range conditions {
		sb.WriteString(field.String())
	}
	return sb.String()
}

func bfsEvaluate() {
	queue := deque.New[*searchNode](16)
	queue.PushBack(&searchNode{})
	for queue.Len() != 0 {
		curr := queue.PopFront()
		//The number of conditions denotes the depth.
		//If we have more conditions than the depth then we can stop
		if len(curr.conditions) == Depth {
			continue
		}
		processStringTargets(curr, queue)
		processIntTargets(curr, queue)
	}
}

func processStringTargets(curr *searchNode, queue *deque.Deque[*searchNode]) {
	for targetIdx := curr.stringTargetStartIdx; targetIdx < len(stringTargets); targetIdx++ {
		targetToAdd := stringTargets[targetIdx]
		for targetValue, count := range targetToAdd.values {
			if !hasSupport(count, len(houses)) {
				continue
			}
			newConditions := getConditionCopy(curr.conditions)
			newConditions = append(newConditions, &searchCondition{
				isString:         true,
				fieldName:        targetToAdd.fieldName,
				fieldValueString: targetValue,
				inequality:       Equal,
			})
			nextNode := &searchNode{
				conditions:           newConditions,
				stringTargetStartIdx: targetIdx + 1,
				intTargetStartIdx:    curr.intTargetStartIdx,
			}
			nextNode.score = calculateCorrelation(nextNode.conditions)
			searchRes[nextNode.conditions.String()] = nextNode
			queue.PushBack(nextNode)
		}
	}
}

func processIntTargets(curr *searchNode, queue *deque.Deque[*searchNode]) {
	for targetIdx := curr.intTargetStartIdx; targetIdx < len(intTargets); targetIdx++ {
		targetToAdd := intTargets[targetIdx]
		//go from start to end accumulating count and only branch out when
		//count is greater than minimum support
		frequency := 0
		for i := targetToAdd.minVal + 1; i <= targetToAdd.maxVal; i += 2 {
			frequency += targetToAdd.counter[i] + targetToAdd.counter[i-1]
			if !hasSupport(frequency, len(houses)) {
				continue
			}
			//less than equal
			newConditions := getConditionCopy(curr.conditions)
			newConditions = append(newConditions, &searchCondition{
				fieldName:     targetToAdd.fieldName,
				fieldValueInt: i,
				inequality:    LessThanEqual,
				isString:      false,
			})
			nextNode := &searchNode{
				intTargetStartIdx:    targetIdx + 1,
				stringTargetStartIdx: curr.stringTargetStartIdx,
				conditions:           newConditions,
			}
			nextNode.score = calculateCorrelation(newConditions)
			queue.PushBack(nextNode)
			searchRes[nextNode.conditions.String()] = nextNode
		}
		//go from end to start and only branch out when
		//count is greater than minimum support
		frequency = 0
		for i := targetToAdd.maxVal - 1; i >= targetToAdd.minVal; i-- {
			frequency += targetToAdd.counter[i] + targetToAdd.counter[i+1]
			if !hasSupport(frequency, len(houses)) {
				continue
			}
			newConditions := getConditionCopy(curr.conditions)
			newConditions = append(newConditions, &searchCondition{
				fieldName:     targetToAdd.fieldName,
				fieldValueInt: i,
				inequality:    GreaterThanEqual,
				isString:      false,
			})
			nextNode := &searchNode{
				intTargetStartIdx:    targetIdx + 1,
				stringTargetStartIdx: curr.stringTargetStartIdx,
				conditions:           newConditions,
			}
			nextNode.score = calculateCorrelation(newConditions)
			queue.PushBack(nextNode)
			searchRes[nextNode.conditions.String()] = nextNode
		}
	}
}

func getConditionCopy[T any](slice []T) []T {
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
