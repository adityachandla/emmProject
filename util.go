package main

import (
	"math"
	"reflect"
)

func Check(err error) {
	if err != nil {
		panic(err)
	}
}

func calculateCorrelation(filterConditions []*searchCondition) float64 {
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
	return numerator / denominator
}

func shouldEvaluate(h *HouseInfo, conditions searchConditions) bool {
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

func isConditionValid(h *HouseInfo, condition *searchCondition) bool {
	value := reflect.ValueOf(h).Elem().FieldByName(condition.fieldName)
	if condition.isString {
		return value.String() == condition.fieldValueString
	}
	if condition.inequality == Equal {
		return value.Int() == int64(condition.fieldValueInt)
	} else if condition.inequality == GreaterThanEqual {
		return value.Int() >= int64(condition.fieldValueInt)
	}
	return value.Int() <= int64(condition.fieldValueInt)
}
