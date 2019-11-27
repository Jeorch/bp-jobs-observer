package models

import "gopkg.in/mgo.v2/bson"

//"_id" : ObjectId("5dd5223d83de972f084b000d"),
//    "fileName" : "171128辉瑞-紫杉醇(白蛋白结合型)1707-1709底层检索(反馈).xlsx",
//    "extension" : "xlsx",
//    "uploaded" : 1574249021827.0,
//    "size" : NumberInt(-1),
//    "url" : "5c922467-d1d3-4303-8900-122259e77aa5/1574249019572",
//    "__v" : NumberInt(0)
type BpFile struct {
	Id        bson.ObjectId `json:"-" bson:"_id"`
	FileName  string        `json:"fileName" bson:"fileName"`
	Extension string        `json:"extension"`
	Uploaded  float64       `json:"uploaded"`
	Size      int           `json:"size"`
	Url       string        `json:"url"`
	V         int           `json:"__v" bson:"__v"`
}
