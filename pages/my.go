package pages

import (
	_ "encoding/json"
)

type Any2 = any

const name = ""

type TagList = []uint

type TagMap = map[string]string

var MyList = make(map[string]string)

type XxMap = map[string]Locations

type Location struct {
	Street  string
	ZipCode string
}

type Locations []Location

// package {pkg}
// func _abc() {
//    const jsPropVar = goStruct{}
//    println(jsPropVar)
// }

func __abc() {
	locs := make(Locations, 0)

	println(locs[0].Street)

	//user := make(TagMap)

	//println(user["x"].Street)
	//
	//json.Unmarshal()
	//var x = make(TagMap)
	//x["3"] = "4"
}

type User2 struct {
	Name2     string
	Locations []Location
	Hello     Location
	Loc       map[string]Location
}

type UsersPage struct {
	Title string
	Users []User2
}

type My struct {
	Name string
	Tags []string
}
