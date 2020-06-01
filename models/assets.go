package models

import "gopkg.in/mgo.v2/bson"

type BpAsset struct {
	Id            bson.ObjectId `json:"-" bson:"_id"`
	Name          string        `json:"name" bson:"name"`
	Description   string        `json:"description" bson:"description"`
	Owner         string        `json:"owner" bson:"owner"`
	Accessibility string        `json:"accessibility" bson:"accessibility"`
	Version       int32         `json:"version" bson:"version"`
	IsNewVersion  bool          `json:"isNewVersion" bson:"isNewVersion"`
	DataType      string        `json:"dataType" bson:"dataType"`
	Providers     []string      `json:"providers" bson:"providers"`
	Markets       []string      `json:"markets" bson:"markets"`
	Molecules     []string      `json:"molecules" bson:"molecules"`
	DataCover     []string      `json:"dataCover" bson:"dataCover"`
	GeoCover      []string      `json:"geoCover" bson:"geoCover"`
	Labels        []string      `json:"labels" bson:"labels"`
	CreateTime    float32       `json:"createTime" bson:"createTime"`
	File          bson.ObjectId `json:"file" bson:"file"`
	Dfs           []interface{} `json:"dfs" bson:"dfs"`
	MartTags      []string      `json:"martTags" bson:"martTags"`
	V             int32         `json:"__v" bson:"__v"`
}
