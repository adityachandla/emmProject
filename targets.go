package main

import "reflect"

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

func calculateStringTargets() []*stringTarget {
	stringTargets := make([]*stringTarget, len(stringFields))
	for idx, field := range stringFields {
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

func calculateIntTargets() []*intTarget {
	intTargets := make([]*intTarget, len(intFields))
	for idx, field := range intFields {
		intTargets[idx] = &intTarget{
			fieldName: field,
			counter:   make(map[int]int),
			minVal:    1 << 30,
			maxVal:    -1,
		}
	}

	for _, house := range houses {
		for idx := range intTargets {
			fieldVal := int(reflect.ValueOf(house).Elem().FieldByName(intTargets[idx].fieldName).Int())
			intTargets[idx].counter[fieldVal]++
			if fieldVal < intTargets[idx].minVal {
				intTargets[idx].minVal = fieldVal
			}
			if fieldVal > intTargets[idx].maxVal {
				intTargets[idx].maxVal = fieldVal
			}
		}
	}
	return intTargets
}
