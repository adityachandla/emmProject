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
	n := float64(len(houses))

	var sumX float64
	var sumY float64

	var squareSumX float64
	var squareSumY float64

	var sumProduct float64
	for _, house := range houses {
		if !shouldEvaluate(house, filterConditions) {
			continue
		}
		count++
		x := house.BuildingArea
		y := float64(house.Price)
		sumX += x
		squareSumX += x * x

		sumY += y
		squareSumY += y * y

		sumProduct += x * y
	}
	numerator := (n * sumProduct) - (sumX * sumY)
	denominator := math.Sqrt((n*squareSumX - (sumX * sumX)) *
		(n*squareSumY - (sumY * sumY)))
	return numerator / denominator, count
}

func shouldEvaluate(h *reader.HouseInfo, conditions Conditions) bool {
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
