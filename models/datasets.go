package models

import (
	"gopkg.in/mgo.v2/bson"
)

const collection string = "datasets"

type BpDataset struct {
	Id          bson.ObjectId   `json:"-" bson:"_id"`
	Parent      []bson.ObjectId `json:"-" bson:"parent"`
	ColNames    []string        `json:"colNames" bson:"colNames"`
	Length      int32           `json:"length"`
	TabName     string          `json:"tabName" bson:"tabName"`
	Url         string          `json:"url"`
	Description string          `json:"description"`
	Status      string          `json:"status"`
	Job         bson.ObjectId   `json:"job"`
	V           int32           `json:"__v" bson:"__v"`
}
