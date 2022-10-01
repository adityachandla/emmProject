package reader

import (
	"bufio"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
)

var tagToFieldNameMapping map[string]string

type HouseInfo struct {
	Suburb         string  `csv:"Suburb"`
	Rooms          int     `csv:"Rooms"`
	HouseType      string  `csv:"Type"`
	Price          int     `csv:"Price"`
	Method         string  `csv:"Method"`
	SellerName     string  `csv:"SellerG"`
	DistFromCenter float64 `csv:"Distance"`
	Bedrooms       int     `csv:"Bedroom2"`
	Bathrooms      int     `csv:"Bathroom"`
	Car            int     `csv:"Car"`
	LandSize       float64 `csv:"Landsize"`
	BuildingArea   float64 `csv:"BuildingArea"`
	YearBuilt      int     `csv:"YearBuilt"`
	CouncilArea    string  `csv:"CouncilArea"`
	Region         string  `csv:"Regionname"`
	PropertyCount  string  `csv:"Propertycount"`
}

func initMapping() {
	tagToFieldNameMapping = make(map[string]string)
	fields := reflect.VisibleFields(reflect.TypeOf(HouseInfo{}))
	for _, field := range fields {
		csvName := field.Tag.Get("csv")
		tagToFieldNameMapping[csvName] = field.Name
	}
}

func ReadHouses() []*HouseInfo {
	initMapping()
	file, _ := os.Open("melb_data.csv")
	reader := bufio.NewReader(file)

	line, _, _ := reader.ReadLine()
	headers := strings.Split(string(line), ",")

	houses := make([]*HouseInfo, 0, 100)

	line, _, err := reader.ReadLine()
	for err == nil {
		entry := strings.Split(string(line), ",")
		HouseInfo := parseInfo(entry, headers)
		houses = append(houses, HouseInfo)
		line, _, err = reader.ReadLine()
	}
	return houses
}

func parseInfo(entry []string, headers []string) *HouseInfo {
	if len(entry) != len(headers) {
		fmt.Println(entry)
		panic("Invalid row encountered")
	}

	info := &HouseInfo{}
	for idx, header := range headers {
		name, exists := tagToFieldNameMapping[header]
		if exists && entry[idx] != "" {
			fieldVal := reflect.ValueOf(info).Elem().FieldByName(name)
			kind := fieldVal.Kind().String()
			if kind == "string" {
				fieldVal.SetString(entry[idx])
			} else if kind == "int" {
				val, err := strconv.ParseFloat(entry[idx], 64)
				check(err)
				fieldVal.SetInt(int64(val))
			} else if kind == "float64" {
				val, err := strconv.ParseFloat(entry[idx], 64)
				check(err)
				fieldVal.SetFloat(val)
			}
		}
	}
	return info
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
