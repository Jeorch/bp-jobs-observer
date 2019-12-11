package models

import "gopkg.in/mgo.v2/bson"

type BpFile struct {
	Id        bson.ObjectId `json:"-" bson:"_id"`
	FileName  string        `json:"fileName" bson:"fileName"`
	Extension string        `json:"extension"`
	Uploaded  float64       `json:"uploaded"`
	Size      int           `json:"size"`
	Url       string        `json:"url"`
	V         int           `json:"__v" bson:"__v"`
}
