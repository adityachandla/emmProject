// Package search
// This package is used for doing a breadth first search on the conditions in our dataset.
// In this file, we define the strategy for doing the breadth first search.
package search

import (
	"container/heap"
	"fmt"
	"github.com/adityachandla/emmTrial/reader"
	"github.com/gammazero/deque"
	"log"
	"math"
	"sync"
)

const (
	Depth      int     = 3
	MinSupport float64 = 0.01
	MaxHeight  int     = 1
	MaxLen     int     = 10
)

var (
	processedCounter int
	houses           []*reader.HouseInfo

	//Target information
	stringTargets []*StringTarget
	intTargets    []*IntTarget

	scoreHeap         *ScoreHeap          //Heap for top MaxLen nodes
	subgroupScoreHeap *SubgroupScoreHeap  // Heap for storing score within subgroups
	baseScore         float64             //Correlation in the entire dataset
	seenNodes         map[string]struct{} //Nodes already seen
	maxHeightNodes    map[string]*Node    //Nodes that are at height less than MaxHeight
	queue             *deque.Deque[*Node] //BFS queue
	mutex             sync.Mutex          //mutex to avoid race while adding to maps and queues
)

func initializeGlobalVariables(h []*reader.HouseInfo) {
	houses = h
	baseScore, _ = CalculateCorrelation(houses, nil)
	log.Printf("Base score is %f", baseScore)

	stringTargets = calculateStringTargets(houses)
	intTargets = calculateIntTargets(houses)

	scoreHeap = &ScoreHeap{}
	subgroupScoreHeap = &SubgroupScoreHeap{}
	seenNodes = make(map[string]struct{})
	maxHeightNodes = make(map[string]*Node)

	queue = deque.New[*Node](16)
}

func BfsEvaluate(h []*reader.HouseInfo) (*ScoreHeap, *SubgroupScoreHeap) {
	initializeGlobalVariables(h)
	//Start with no conditions
	queue.PushBack(&Node{})
	for queue.Len() != 0 {
		curr := queue.PopFront()
		//The number of Conditions denotes the depth.
		//If we have more Conditions than the depth then we can stop
		if len(curr.Conditions) == Depth {
			continue
		}
		wg := &sync.WaitGroup{}
		wg.Add(2)
		go func() {
			defer wg.Done()
			processStringTargets(curr)
		}()
		//We process strings in both increasing and decreasing order
		//So if we have 1,2,3 for a field then less than equal
		//conditions would be <=1, <=2, <=3
		go func() {
			defer wg.Done()
			processIntTargetsLessThanEqual(curr)
		}()
		//and greater than equal conditions would be >=1, >=2,>=3
		processIntTargetsGreaterThanEqual(curr)
		wg.Wait()
	}
	return scoreHeap, subgroupScoreHeap
}

func processStringTargets(curr *Node) {
	for targetIdx := curr.stringTargetStartIdx; targetIdx < len(stringTargets); targetIdx++ {
		targetToAdd := stringTargets[targetIdx]
		for targetValue, count := range targetToAdd.Values {
			//This is just a preemptive pruning because we don't consider the fact that curr will already
			//be a subset of all the houses. The actual pruning happens after we have the score and count.
			if !hasSupport(count, len(houses)) {
				continue
			}
			newCondition := getStringEqualityCondition(targetToAdd.FieldName, targetValue)
			nextNode := getNodeWithAddedCondition(newCondition, curr)
			nextNode.stringTargetStartIdx = targetIdx + 1
			processNode(nextNode)
		}
	}
}

func getStringEqualityCondition(name, value string) *Condition {
	return &Condition{
		IsString:         true,
		FieldName:        name,
		FieldValueString: value,
		Inequality:       Equal,
	}
}

func processIntTargetsLessThanEqual(curr *Node) {
	for targetIdx := curr.intTargetStartIdx; targetIdx < len(intTargets); targetIdx++ {
		targetToAdd := intTargets[targetIdx]
		//go from start to end accumulating count and only branch out when
		//count is greater than minimum support
		frequency := 0
		for i := targetToAdd.MinVal + 1; i <= targetToAdd.MaxVal-1; i += 2 {
			frequency += targetToAdd.Counter[i] + targetToAdd.Counter[i-1]
			if !hasSupport(frequency, len(houses)) {
				continue
			}
			newCondition := getIntLessThanEqualCondition(targetToAdd.FieldName, i)
			nextNode := getNodeWithAddedCondition(newCondition, curr)
			nextNode.intTargetStartIdx = targetIdx + 1
			processNode(nextNode)
		}
	}
}

func getIntLessThanEqualCondition(name string, value int) *Condition {
	return &Condition{
		FieldName:     name,
		FieldValueInt: value,
		Inequality:    LessThanEqual,
		IsString:      false,
	}
}

func processIntTargetsGreaterThanEqual(curr *Node) {
	for targetIdx := curr.intTargetStartIdx; targetIdx < len(intTargets); targetIdx++ {
		targetToAdd := intTargets[targetIdx]
		//go from end to start and only branch out when
		//count is greater than minimum support
		frequency := 0
		for i := targetToAdd.MaxVal - 1; i >= targetToAdd.MinVal+1; i-- {
			frequency += targetToAdd.Counter[i] + targetToAdd.Counter[i+1]
			if !hasSupport(frequency, len(houses)) {
				continue
			}
			newCondition := getIntGreaterThanEqualCondition(targetToAdd.FieldName, i)
			nextNode := getNodeWithAddedCondition(newCondition, curr)
			nextNode.intTargetStartIdx = targetIdx + 1
			processNode(nextNode)
		}
	}
}

func getIntGreaterThanEqualCondition(name string, value int) *Condition {
	return &Condition{
		FieldName:     name,
		FieldValueInt: value,
		Inequality:    GreaterThanEqual,
		IsString:      false,
	}
}

func processNode(nextNode *Node) {
	nextNode.Correlation, nextNode.Size = CalculateCorrelation(houses, nextNode.Conditions)
	nextNode.Score = math.Abs(nextNode.Correlation - baseScore)
	if hasSupport(nextNode.Size, len(houses)) {
		evaluateSubgroupScore(nextNode)
		addNode(nextNode)
	}
}

func evaluateSubgroupScore(nextNode *Node) {
	for _, condition := range nextNode.Conditions {
		node, contains := maxHeightNodes[condition.String()]
		if !contains {
			continue
		}
		correlationDiff := math.Abs(node.Correlation - nextNode.Correlation)
		if correlationDiff > nextNode.SubgroupScore {
			nextNode.SubgroupScore = correlationDiff
			nextNode.SubgroupWithin = condition.String()
		}
	}
}

func addNode(node *Node) {
	mutex.Lock()
	defer mutex.Unlock()
	if _, present := seenNodes[node.Conditions.String()]; present {
		return
	}

	if len(node.Conditions) <= MaxHeight {
		maxHeightNodes[node.Conditions.String()] = node
	}
	queue.PushBack(node)
	addToScoreHeap(node)
	addToSubgroupHeap(node)
	seenNodes[node.Conditions.String()] = struct{}{}

	fmt.Printf("Added nodes : %d\r", processedCounter)
	processedCounter++
}

func addToScoreHeap(node *Node) {
	if scoreHeap.Len() < MaxLen {
		heap.Push(scoreHeap, node)
	} else if node.Score > (*scoreHeap)[0].Score {
		heap.Pop(scoreHeap)
		heap.Push(scoreHeap, node)
	}
}

func addToSubgroupHeap(node *Node) {
	if node.SubgroupScore == 0 {
		return
	}
	if subgroupScoreHeap.Len() < MaxLen {
		heap.Push(subgroupScoreHeap, node)
	} else if node.SubgroupScore > (*subgroupScoreHeap)[0].SubgroupScore {
		heap.Pop(subgroupScoreHeap)
		heap.Push(subgroupScoreHeap, node)
	}
}
