package detail

import (
	"context"
	"fmt"
	"github.com/PharbersDeveloper/bp-go-lib/log"
	"github.com/PharbersDeveloper/bp-jobs-observer/models"
	"github.com/PharbersDeveloper/bp-jobs-observer/utils"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"regexp"
	"strings"
	"time"
)

type ObserverInfo struct {
	Id             string                 `json:"id"`
	DBHost         string                 `json:"db_host"`
	DBPort         string                 `json:"db_port"`
	Database       string                 `json:"database"`
	Collection     string                 `json:"collection"`
	Conditions     map[string]interface{} `json:"conditions"`
	ParallelNumber int                    `json:"parallel_number"`
	DBUser         string                 `json:"db_user"`
	DBPass         string                 `json:"db_pass"`
}

var (
	dbSession    *mgo.Session
	jobChan      chan models.BpDataset
	s3Svc  		*s3.S3
)

func (observer *ObserverInfo) Open() {
	sess, err := mgo.Dial(fmt.Sprintf("%s:%s", observer.DBHost, observer.DBPort))
	if err != nil {
		log.NewLogicLoggerBuilder().Build().Error(err.Error())
	}
	cred := mgo.Credential{
		Username: observer.DBUser,
		Password: observer.DBPass,
	}
	err = sess.Login(&cred)
	if err != nil {
		log.NewLogicLoggerBuilder().Build().Error(err.Error())
		if sess != nil {
			sess.Refresh()
		}
	}
	if sess != nil {
		sess.SetMode(mgo.Monotonic, true)
		dbSession = sess
	}

	if err != nil {
		log.NewLogicLoggerBuilder().Build().Error(err.Error())
	}

	s3Svc = s3.New(session.New())

}

func (observer *ObserverInfo) Exec() {
	execLogger := log.NewLogicLoggerBuilder().SetTraceId(observer.Id).Build()
	execLogger.Info("start exec")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//查询Jobs
	jobs, err := observer.queryJobs()
	if err != nil {
		execLogger.Error(err.Error())
	}

	length := len(jobs)
	execLogger.Info("jobs length=", length)

	if length > 0 {
		jobChan = make(chan models.BpDataset, observer.ParallelNumber)

		//分配worker执行Job
		for id := 1; id <= observer.ParallelNumber; id++ {
			go observer.worker(id, jobChan, ctx)
		}

		//将所有 job push 到 chan队列
		pushJobs(jobChan, jobs)
	}

	execLogger.Info("Ended!")
	time.Sleep(10 * time.Second) //10秒钟后ctx局部变量done
}

func (observer *ObserverInfo) Close() {
	if jobChan != nil {
		close(jobChan)
	}
}

func (observer *ObserverInfo) queryJobs() ([]models.BpDataset, error) {

	logger := log.NewLogicLoggerBuilder().SetTraceId(observer.Id).Build()
	logger.Info("query jobs from db")

	var datasets []models.BpDataset
	//err := dbSession.DB(observer.Database).C(observer.Collection).Find(bson.M{}).All(&datasets)
	err := dbSession.DB(observer.Database).C(observer.Collection).Find(observer.Conditions).All(&datasets)
	if err != nil {
		return nil, err
	}

	var allMart []models.BpMart
	err = dbSession.DB(observer.Database).C("marts").Find(bson.M{}).All(&allMart)
	if err != nil {
		return nil, err
	}
	oldDfsIds, err := observer.getAllMartDfsParents(allMart)
	if err != nil {
		return nil, err
	}

	jobs := make([]models.BpDataset, 0)
	for _, dfs := range datasets {
		if isNotOldDfs(dfs.Id, oldDfsIds) {
			jobs = append(jobs, dfs)
		}
	}

	return jobs, nil
}

func (observer *ObserverInfo) getAllMartDfsParents(allMart []models.BpMart) ([]bson.ObjectId, error) {
	var res []bson.ObjectId
	for _, mart := range allMart {
		for _, id := range mart.Dfs {
			var ds models.BpDataset
			err := dbSession.DB(observer.Database).C(observer.Collection).Find(bson.M{"_id": id}).One(&ds)
			if err != nil {
				return nil, err
			}
			if len(ds.Parent) != 0 {
				res = append(res, ds.Parent...)
			}
		}
	}
	return res, nil
}

func isNotOldDfs(id bson.ObjectId, oldDfsIds []bson.ObjectId) bool {
	for _, oldDfsId := range oldDfsIds {
		if id.Hex() == oldDfsId.Hex() {
			return false
		}
	}
	return true
}

func pushJobs(jobChan chan<- models.BpDataset, jobs []models.BpDataset) {
	logger := log.NewLogicLoggerBuilder().Build()
	logger.Info("push jobs to chan")
	for _, job := range jobs {
		jobChan <- job
	}
}

func (observer *ObserverInfo) worker(id int, jobChan <-chan models.BpDataset, ctx context.Context) {
	workerLogger := log.NewLogicLoggerBuilder().Build()
	workerLogger.Info("worker", id, " standby!")

	for {
		select {
		case <-ctx.Done():
			workerLogger.Infof("worker-%d stop", id)
			return
		case j := <-jobChan:
			{

				jobLogger := log.NewLogicLoggerBuilder().Build()
				jobLogger.Infof("worker-%d start job=%v", id, j.Id.Hex())

				//checkDatasetHasErrorFileInS3
				dfsHasErr, err := isDatasetHasErrorFileInS3(j.Url)
				if err != nil {
					jobLogger.Errorf("worker-%d error in job=%v, msg=(%v)", id, j.Id.Hex(), err.Error())
					return
				}

				if dfsHasErr {
					//删除dfs的error结果在s3上
					go deleteErrorResultInS3(j.Url)
					//删除本条dfs数据在datasets表上
					go observer.deleteDatasetById(j.Id)
					//找到包含本条dfsId的assets数据，清空其dfs数组
					go observer.resetAssetsDfsByDatasetId(j.Id)

				}

				jobLogger.Infof("worker-%d sanded job=%v", id, j.Id.Hex())
			}
		}
	}

}

func isDatasetHasErrorFileInS3(dfsUrl string) (bool, error) {

	hasErr := false
	logger := log.NewLogicLoggerBuilder().Build()

	s3Path := strings.Replace(dfsUrl,utils.S3_PathPrefix, "", 1)
	bucketAndPath := strings.SplitN(s3Path, "/", 2)
	if len(bucketAndPath) != 2 {
		logger.Error("Url's s3a path error format.")
	}
	bucket := bucketAndPath[0]
	contentsPath := bucketAndPath[1]
	errPath := utils.Reverse(strings.Replace(utils.Reverse(contentsPath), utils.Reverse("contents"), utils.Reverse("err"), 1))

	objs, err := utils.S3_ListObjects(s3Svc, bucket, errPath)
	if err != nil {
		logger.Error(err.Error())
		return hasErr, err
	}

	if len(objs) == 0 {
		return hasErr, nil
	}

	for _, obj := range objs {
		match, _ := regexp.MatchString("(.*)_SUCCESS ", *obj.Key)
		if !match && !utils.S3_IsEmptyObject(obj) {
			hasErr = true
		}
	}
	return hasErr, nil

}

func deleteErrorResultInS3(dfsUrl string) {
	logger := log.NewLogicLoggerBuilder().Build()

	s3Path := strings.Replace(dfsUrl,utils.S3_PathPrefix, "", 1)
	bucketAndPath := strings.SplitN(s3Path, "/", 2)
	if len(bucketAndPath) != 2 {
		logger.Error("Url's s3a path error format.")
	}
	bucket := bucketAndPath[0]
	contentsPath := bucketAndPath[1]
	parentPath := utils.Reverse(strings.Replace(utils.Reverse(contentsPath), utils.Reverse("contents"), "", 1))

	err := utils.S3_DeleteFolder(s3Svc, bucket, parentPath)
	if err != nil {
		logger.Error(err.Error())
	}

}

func (observer *ObserverInfo) deleteDatasetById(id bson.ObjectId)  {
	logger := log.NewLogicLoggerBuilder().Build()
	err := dbSession.DB(observer.Database).C(observer.Collection).RemoveId(id)
	if err != nil {
		logger.Error(err.Error())
	}
}

func (observer *ObserverInfo) resetAssetsDfsByDatasetId(datasetId bson.ObjectId)  {
	logger := log.NewLogicLoggerBuilder().Build()

	var asset models.BpAsset
	err := dbSession.DB(observer.Database).C("assets").Find(bson.M{"dfs":bson.M{"$in":[]bson.ObjectId{datasetId}}}).One(&asset)
	if err != nil {
		logger.Error(err.Error())
	}
	fmt.Println(asset)

	if len(asset.Dfs) != 0 {
		asset.Dfs = nil
		err = dbSession.DB(observer.Database).C("assets").UpdateId(asset.Id, asset)
		if err != nil {
			logger.Error(err.Error())
		}
		fmt.Println(asset)
	}
}
