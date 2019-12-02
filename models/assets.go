package models

import "gopkg.in/mgo.v2/bson"

//"_id" : ObjectId("5dd5223f83de972f084b000e"),
//    "traceId" : "5c922467-d1d3-4303-8900-122259e77aa5",
//    "name" : "171128辉瑞-紫杉醇(白蛋白结合型)1707-1709底层检索(反馈).xlsx",
//    "description" : "L:/全原盘数据/D盘备份/Pharbers文件/Pfizer/CPA国药数据/1709/171128辉瑞-紫杉醇(白蛋白结合型)1707-1709底层检索(反馈).xlsx",
//    "owner" : "auto robot",
//    "accessibility" : "w",
//    "version" : NumberInt(0),
//    "isNewVersion" : true,
//    "dataType" : "file",
//    "providers" : [
//        "Pfizer"
//    ],
//    "markets" : [
//
//    ],
//    "molecules" : [
//
//    ],
//    "dataCover" : [
//        "201701",
//        "201709"
//    ],
//    "geoCover" : [
//
//    ],
//    "labels" : [
//        "其他"
//    ],
//    "file" : ObjectId("5dd5223d83de972f084b000d"),
//    "dfs" : [
//
//    ],
//    "__v" : NumberInt(0)
//}
//{
//    "_id" : ObjectId("5dd5224183de972f084b0018"),
//    "traceId" : "f4fd40a4-197f-42dc-a635-214e5ee63a6f",
//    "name" : "2015年1月-2017年4月湖北省百普乐产品数据.xlsx",
//    "description" : "L:/全原盘数据/D盘备份/Pharbers文件/Pfizer/CPA国药数据/1704/2015年1月-2017年4月湖北省百普乐产品数据.xlsx",
//    "owner" : "auto robot",
//    "accessibility" : "w",
//    "version" : NumberInt(0),
//    "isNewVersion" : true,
//    "dataType" : "file",
//    "providers" : [
//        "Pfizer"
//    ],
//    "markets" : [
//
//    ],
//    "molecules" : [
//
//    ],
//    "dataCover" : [
//        "201501",
//        "201704"
//    ],
//    "geoCover" : [
//
//    ],
//    "labels" : [
//        "原始数据"
//    ],
//    "file" : ObjectId("5dd5224083de972f084b0013"),
//    "dfs" : [
//
//    ],
//    "__v" : NumberInt(0)

type BpAsset struct {
	Id          bson.ObjectId `json:"-" bson:"_id"`
	TraceId     string        `json:"traceId" bson:"traceId"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Providers   []string      `json:"providers" bson:"providers"`
	Markets     []string      `json:"markets" bson:"markets"`
	Molecules   []string      `json:"molecules" bson:"molecules"`
	DataCover   []string      `json:"dataCover" bson:"dataCover"`
	GeoCover    []string      `json:"geoCover" bson:"geoCover"`
	Labels      []string      `json:"labels"`
	File        bson.ObjectId `json:"file"`
	Dfs         []interface{} `json:"dfs"`
}
