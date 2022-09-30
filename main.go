package main

import (
	"fmt"
	"github.com/gammazero/deque"
	"log"
	"reflect"
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

type intTarget struct {
	fieldName string
	counter   map[int]int
	minVal    int
	maxVal    int
}

type stringTarget struct {
	fieldName string
	values    map[string]int
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

var searchRes = make([]*searchNode, 0, 100)

func main() {
	houses := ReadHouses()
	//intFields := []string{"Bedrooms", "Bathrooms", "Car", "Rooms"}
	stringFields := []string{"HouseType", "Suburb"}

	//intTargets := calculateIntTargets(houses, intFields)
	stringTargets := calculateStringTargets(houses, stringFields)
	bfsEvaluate(stringTargets, houses)
	for _, res := range searchRes {
		log.Printf("%s %f\n", res.conditions, res.score)
	}
	fmt.Printf("Comparision to: %f", calculateCorrelation(houses, nil))
}

func bfsEvaluate(targets []*stringTarget, houses []*HouseInfo) {
	queue := deque.New[*searchNode](16) //TODO can we make this dynamic?
	queue.PushBack(&searchNode{})
	for queue.Len() != 0 {
		curr := queue.PopFront()
		for targetIdx := curr.stringTargetStartIdx; targetIdx < len(targets); targetIdx++ {
			targetToAdd := targets[targetIdx]
			for targetValue, count := range targetToAdd.values {
				if !hasSupport(count, len(houses)) {
					continue
				}
				var newConditions []*searchCondition
				if curr.conditions == nil {
					newConditions = make([]*searchCondition, 0, 4) //TODO This should be max depth
				} else {
					newConditions = copyOf(curr.conditions)
				}
				newConditions = append(newConditions, getCondition(targetValue, targetToAdd))
				nextNode := &searchNode{
					conditions:           newConditions,
					stringTargetStartIdx: targetIdx + 1,
				}
				nextNode.score = calculateCorrelation(houses, nextNode.conditions)
				searchRes = append(searchRes, nextNode)
				queue.PushBack(nextNode)
			}
		}
	}
}

func hasSupport(count, total int) bool {
	return (float64(count) / float64(total)) > 0.01 //TODO Parameterize this
}

func copyOf[T any](slice []T) []T {
	newSlice := make([]T, len(slice))
	copy(newSlice, slice)
	return newSlice
}

func getCondition(targetValue string, target *stringTarget) *searchCondition {
	return &searchCondition{
		isString:         true,
		fieldName:        target.fieldName,
		fieldValueString: targetValue,
		inequality:       Equal,
	}
}

func calculateStringTargets(houses []*HouseInfo, fields []string) []*stringTarget {
	stringTargets := make([]*stringTarget, len(fields))
	for idx, field := range fields {
		stringTargets[idx] = &stringTarget{fieldName: field, values: make(map[string]int)}
	}

	for _, house := range houses {
		for idx := range stringTargets {
			value := reflect.ValueOf(house).Elem().FieldByName(stringTargets[idx].fieldName).String()
			stringTargets[idx].values[value]++
		}
	}

	return stringTargets
}

func calculateIntTargets(houses []*HouseInfo, fields []string) []*intTarget {
	intTargets := make([]*intTarget, len(fields))
	for idx, field := range fields {
		intTargets[idx] = &intTarget{
			fieldName: field,
			counter:   make(map[int]int),
			minVal:    -1,
			maxVal:    1 << 30,
		}
	}

	for _, house := range houses {
		for idx := range intTargets {
			fieldVal := int(reflect.ValueOf(house).Elem().FieldByName(intTargets[idx].fieldName).Int())
			intTargets[idx].counter[fieldVal]++
			if fieldVal > intTargets[idx].minVal {
				intTargets[idx].minVal = fieldVal
			}
			if fieldVal < intTargets[idx].maxVal {
				intTargets[idx].maxVal = fieldVal
			}
		}
	}
	return intTargets
}
