package observer_oss_task

import (
	"context"
	"github.com/PharbersDeveloper/bp-go-lib/kafka"
	"github.com/PharbersDeveloper/bp-go-lib/log"
	"github.com/PharbersDeveloper/bp-jobs-observer/models"
	"github.com/PharbersDeveloper/bp-jobs-observer/models/record"
	"github.com/PharbersDeveloper/bp-jobs-observer/utils"
	"github.com/hashicorp/go-uuid"
	"gopkg.in/mgo.v2/bson"
	"time"
)


func (observer *ObserverInfo) queryJobs() ([]record.OssTask, error) {

	logger := log.NewLogicLoggerBuilder().Build()
	logger.Info("query jobs from db")

	var assets []models.BpAsset
	//err := dbSession.DB(observer.Database).C(observer.Collection).Find(observer.Conditions).Limit(observer.ParallelNumber).All(&assets)
	err := dbSession.DB(observer.Database).C(observer.Collection).Find(observer.Conditions).All(&assets)
	if err != nil {
		return nil, err
	}

	expired := time.Duration(observer.SingleJobTimeoutSecond) * time.Second

	jobs := make([]record.OssTask, 0)
	for _, asset := range assets {
		file, e := observer.queryFile(asset.File)
		if e != nil {
			return nil, e
		}
		if file.Extension == "xlsx" || file.Extension == "xls" {
			assetId := asset.Id.Hex()

			//TODO:在此处使用redis来check是否存在相同的job内容
			//TODO:暂时使用redis来存储以assetId为job内容的唯一标示
			//TODO:暂时以assetId为job内容的唯一标示
			exist, e := utils.CheckKeyExistInRedis(assetId)
			if e != nil {
				return nil, e
			}
			if !exist {
				count, e := utils.AddKey2Redis(assetId)
				if e != nil {
					return nil, e
				}
				logger.Infof("将assetId=%s的asset加入job队列，redis inc count=%d", assetId, count)

				ok, e := utils.SetKeyExpire(assetId, expired)
				if !ok {
					logger.Infof("设置assetId=%s过期时间失败，尝试重新设置过期时间", assetId)
					_, _ = utils.SetKeyExpire(assetId, expired)
				}

				//TODO:此处为拼接Job
				//TODO:此处使用UUID生成JobId
				newId, err := uuid.GenerateUUID()
				if err != nil {
					return nil, e
				}
				{
					job := record.OssTask{
						TitleIndex: nil,
						JobId:      newId,
						TraceId:    newId,
						AssetId:	assetId,
						OssKey:     file.Url,
						FileType:   file.Extension,
						FileName:   file.FileName,
						SheetName:  "",
						Labels:     asset.Labels,
						DataCover:  asset.DataCover,
						GeoCover:   asset.GeoCover,
						Markets:    asset.Markets,
						Molecules:  asset.Molecules,
						Providers:  asset.Providers,
					}
					jobs = append(jobs, job)
				}

			}
		} else {
			println(file.Extension)
		}

	}

	return jobs, nil
}

func (observer *ObserverInfo) queryFile(id bson.ObjectId) (models.BpFile, error) {
	var file models.BpFile
	err := dbSession.DB(observer.Database).C("files").Find(bson.M{"_id": id}).One(&file)
	if err != nil {
		return file, err
	}
	return file, err
}

func pushJobs(jobChan chan<- record.OssTask, jobs []record.OssTask) {
	logger := log.NewLogicLoggerBuilder().Build()
	logger.Info("push jobs to chan")
	for _, job := range jobs {
		jobChan <- job
	}
}

func (observer *ObserverInfo) worker(id int, jobChan <-chan record.OssTask, ctx context.Context) {
	workerLogger := log.NewLogicLoggerBuilder().Build()
	workerLogger.Info("worker", id, " standby!")

	for {
		select {
		case <-ctx.Done():
			workerLogger.Infof("worker-%d stop", id)
			return
		case j := <-jobChan:
			{
				jobId := j.JobId
				traceId := j.TraceId

				jobLogger := log.NewLogicLoggerBuilder().SetTraceId(traceId).SetJobId(jobId).Build()
				jobLogger.Infof("worker-%d start job=%v", id, j)

				//send job request
				err := sendJobRequest(observer.RequestTopic, j)
				if err != nil {
					jobLogger.Error(err.Error())
				}
				jobLogger.Infof("worker-%d sanded job=%v", id, j)
			}
		}

	}

	workerLogger.Info("worker", id, " done!")
}

func sendJobRequest(topic string, job record.OssTask) error {

	specificRecordByteArr, err := kafka.EncodeAvroRecord(&job)
	if err != nil {
		return err
	}

	err = producer.Produce(topic, []byte(job.TraceId), specificRecordByteArr)
	return err
}

func (observer *ObserverInfo) scheduleJob(jobChan chan record.OssTask, ctx context.Context) {

	//1. 设置ticker 循环时间
	//2. 扫描当前的Jobs，确定job len为空，且jobStatus都为End
	//3. 重新查询Jobs
	//4. 将查询的job push到job chan

	logger := log.NewLogicLoggerBuilder().Build()

	//设置时间周期
	d := time.Duration(observer.ScheduleDurationSecond) * time.Second
	t := time.NewTimer(d)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			//上层（调用协程的）结束，终止子协程
			logger.Info("Stop schedule because context is done.")
			return
		case <-t.C:
			logger.Info("start schedule job ...")

			//查询Jobs
			jobs, err := observer.queryJobs()
			if err != nil {
				logger.Error(err.Error())
			}

			length := len(jobs)
			logger.Info("jobs length=", length)

			pushJobs(jobChan, jobs)

			// need reset timer
			t.Reset(d)
		}

	}

}
