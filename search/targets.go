package search

import (
	"github.com/adityachandla/emmTrial/reader"
	"reflect"
)

type IntTarget struct {
	FieldName string
	Counter   map[int]int
	MinVal    int
	MaxVal    int
}

type StringTarget struct {
	FieldName string
	Values    map[string]int
}

var (
	intFields    = []string{"Bedrooms", "Bathrooms", "Car", "Rooms"}
	stringFields = []string{"HouseType", "Suburb"}
)

// We calculate the various values of string fields as well as the count of
// each value so that it is easier to prune minimum support.
func calculateStringTargets(houses []*reader.HouseInfo) []*StringTarget {
	stringTargets := make([]*StringTarget, len(stringFields))
	for idx, field := range stringFields {
		stringTargets[idx] = &StringTarget{FieldName: field, Values: make(map[string]int)}
	}

	for _, house := range houses {
		for idx := range stringTargets {
			value := reflect.ValueOf(house).Elem().FieldByName(stringTargets[idx].FieldName).String()
			stringTargets[idx].Values[value]++
		}
	}

	return stringTargets
}

// For integer targets, we calculate the frequency of each field as well as their range
// so that we can calculate minimum support without going over the data again.
func calculateIntTargets(houses []*reader.HouseInfo) []*IntTarget {
	intTargets := make([]*IntTarget, len(intFields))
	for idx, field := range intFields {
		intTargets[idx] = &IntTarget{
			FieldName: field,
			Counter:   make(map[int]int),
			MinVal:    1 << 30,
			MaxVal:    -1,
		}
	}

	for _, house := range houses {
		for idx := range intTargets {
			fieldVal := int(reflect.ValueOf(house).Elem().FieldByName(intTargets[idx].FieldName).Int())
			intTargets[idx].Counter[fieldVal]++
			if fieldVal < intTargets[idx].MinVal {
				intTargets[idx].MinVal = fieldVal
			}
			if fieldVal > intTargets[idx].MaxVal {
				intTargets[idx].MaxVal = fieldVal
			}
		}
	}
	return intTargets
}
