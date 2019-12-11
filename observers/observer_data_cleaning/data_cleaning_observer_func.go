package observer_data_cleaning

import (
	"context"
	"github.com/PharbersDeveloper/bp-go-lib/kafka"
	"github.com/PharbersDeveloper/bp-go-lib/log"
	"github.com/PharbersDeveloper/bp-jobs-observer/models"
	"github.com/PharbersDeveloper/bp-jobs-observer/models/record"
	"github.com/PharbersDeveloper/bp-jobs-observer/utils"
	"github.com/hashicorp/go-uuid"
	"time"
)

func (observer *ObserverInfo) queryJobs() ([]record.HiveTask, error) {

	logger := log.NewLogicLoggerBuilder().Build()
	logger.Info("query jobs from db")

	newTraceId, err := uuid.GenerateUUID()

	var datasets []models.BpDataset
	err = dbSession.DB(observer.Database).C(observer.Collection).Find(observer.Conditions).All(&datasets)
	if err != nil {
		return nil, err
	}

	expired := time.Duration(observer.SingleJobTimeoutSecond) * time.Second

	jobs := make([]record.HiveTask, 0)
	for _, dataset := range datasets {
		datasetId := dataset.Id.Hex()

		//TODO:在此处使用redis来check是否存在相同的job内容
		//TODO:暂时使用redis来存储以 datasetId 为job内容的唯一标示
		exist, e := utils.CheckKeyExistInRedis(datasetId)
		if e != nil {
			return nil, e
		}
		if !exist {
			count, e := utils.AddKey2Redis(datasetId)
			if e != nil {
				return nil, e
			}
			logger.Infof("将assetId=%s的asset加入job队列，redis inc count=%d", datasetId, count)

			ok, e := utils.SetKeyExpire(datasetId, expired)
			if !ok {
				logger.Infof("设置assetId=%s过期时间失败，尝试重新设置过期时间", datasetId)
				_, _ = utils.SetKeyExpire(datasetId, expired)
			}

			//TODO:此处为拼接Job
			//TODO:此处使用UUID生成JobId
			newJobId, err := uuid.GenerateUUID()
			if err != nil {
				return nil, e
			}
			{
				job := record.HiveTask{
					JobId:     newJobId,
					TraceId:   newTraceId,
					DatasetId: datasetId,
					Url:       dataset.Url,
					Length:    dataset.Length,
					TaskType:  "create", //TODO：暂时只是创建hive表的请求
					Remarks:   "",       //TODO：备注，预留字段
				}
				jobs = append(jobs, job)
			}

		}

	}

	return jobs, nil
}

func pushJobs(jobChan chan<- record.HiveTask, jobs []record.HiveTask) {
	logger := log.NewLogicLoggerBuilder().Build()
	logger.Info("push jobs to chan")
	for _, file := range jobs {
		jobChan <- file
	}
}

func (observer *ObserverInfo) worker(id int, jobChan <-chan record.HiveTask, ctx context.Context) {
	workerLogger := log.NewLogicLoggerBuilder().Build()
	workerLogger.Info("worker", id, " standby!")

	for {
		select {
		case <-ctx.Done():
			workerLogger.Infof("worker-%d stop", id)
			return
		case j := <-jobChan:
			{
				//TODO: 由于asset是以traceId为区分，jobId未使用，这里自动化Job以traceId作为JobId
				jobId := j.TraceId
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

func sendJobRequest(topic string, job record.HiveTask) error {

	specificRecordByteArr, err := kafka.EncodeAvroRecord(&job)
	if err != nil {
		return err
	}

	err = producer.Produce(topic, []byte(job.TraceId), specificRecordByteArr)
	return err
}

func (observer *ObserverInfo) scheduleJob(jobChan chan record.HiveTask, ctx context.Context) {

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
			files, err := observer.queryJobs()
			if err != nil {
				logger.Error(err.Error())
			}

			length := len(files)
			logger.Info("jobs length=", length)

			pushJobs(jobChan, files)

			// need reset timer
			t.Reset(d)
		}

	}

}
