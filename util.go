package main

import "math"

func Check(err error) {
	if err != nil {
		panic(err)
	}
}

func filter[K any](list []*K, filterExp func(*K) bool) []*K {
	filtered := make([]*K, 0, 10)
	for _, val := range list {
		if filterExp(val) {
			filtered = append(filtered, val)
		}
	}
	return filtered
}

func calculateCorrelation(houses []*HouseInfo, filterConditions []func(*HouseInfo) bool) float64 {
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

func shouldEvaluate(h *HouseInfo, filters []func(*HouseInfo) bool) bool {
	if filters == nil {
		return true
	}
	for _, filter := range filters {
		if !filter(h) {
			return false
		}
	}
	return true
}
