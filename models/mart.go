package models

import "gopkg.in/mgo.v2/bson"

type BpMart struct {
	Id          bson.ObjectId   `json:"-" bson:"_id"`
	Dfs         []bson.ObjectId `json:"dfs" bson:"dfs"`
	Name        string          `json:"name" bson:"name"`
	Url         string          `json:"url" bson:"url"`
	DataType    string          `json:"dataType" bson:"dataType"`
	Description string          `json:"description" bson:"description"`
}
