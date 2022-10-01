// Package search
// This package contains parameters for the search.
package search

import (
	"fmt"
	"sort"
	"strings"
)

// This is used to define the inequality type in our conditions
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
	panic("Unexpected value for Inequality")
}

// Node contains all the fields required to evaluate the score for
// the current node as well as to further branch.
type Node struct {
	Conditions           Conditions
	Score                float64
	stringTargetStartIdx int
	intTargetStartIdx    int
}

// Condition defines the inequality that we need to evaluate for the node.
type Condition struct {
	FieldName        string
	IsString         bool //it will either be string or int for now
	FieldValueString string
	FieldValueInt    int
	Inequality       inequality
}

func (cond *Condition) String() string {
	if cond.IsString {
		return fmt.Sprintf("(%s %s %s)", cond.FieldName, cond.Inequality, cond.FieldValueString)
	}
	return fmt.Sprintf("(%s %s %d)", cond.FieldName, cond.Inequality, cond.FieldValueInt)
}

type Conditions []*Condition

func (conditions Conditions) Len() int { return len(conditions) }
func (conditions Conditions) Less(i, j int) bool {
	return conditions[i].FieldName < conditions[j].FieldName
}
func (conditions Conditions) Swap(i, j int) {
	conditions[i], conditions[j] = conditions[j], conditions[i]
}

func (conditions Conditions) String() string {
	sort.Sort(conditions)
	sb := strings.Builder{}
	for _, field := range conditions {
		sb.WriteString(field.String())
	}
	return sb.String()
}
