package search

import (
	"github.com/adityachandla/emmTrial/reader"
	"math"
	"reflect"
)

// CalculateCorrelation is a single pass correlation calculation
// It is a modified version of https://www.geeksforgeeks.org/program-find-correlation-coefficient/
func CalculateCorrelation(houses []*reader.HouseInfo, filterConditions []*Condition) (float64, int) {
	count := 0

	var sumX float64
	var sumY float64

	var squareSumX float64
	var squareSumY float64

	var sumProduct float64
	for _, house := range houses {
		if !shouldEvaluateHouse(house, filterConditions) {
			continue
		}
		count++
		x := house.LandSize
		y := float64(house.Price)
		sumX += x
		squareSumX += x * x

		sumY += y
		squareSumY += y * y

		sumProduct += x * y
	}
	numerator := (float64(count) * sumProduct) - (sumX * sumY)
	denominator := math.Sqrt((float64(count)*squareSumX - (sumX * sumX)) *
		(float64(count)*squareSumY - (sumY * sumY)))
	return numerator / denominator, count
}

func shouldEvaluateHouse(h *reader.HouseInfo, conditions Conditions) bool {
	if conditions == nil {
		return true
	}
	for _, condition := range conditions {
		if !isConditionValid(h, condition) {
			return false
		}
	}
	return true
}

func isConditionValid(h *reader.HouseInfo, condition *Condition) bool {
	value := reflect.ValueOf(h).Elem().FieldByName(condition.FieldName)
	if condition.IsString {
		return value.String() == condition.FieldValueString
	}
	if condition.Inequality == Equal {
		return value.Int() == int64(condition.FieldValueInt)
	} else if condition.Inequality == GreaterThanEqual {
		return value.Int() >= int64(condition.FieldValueInt)
	}
	return value.Int() <= int64(condition.FieldValueInt)
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
