package observer_traceback

import (
	"bufio"
	"context"
	"github.com/PharbersDeveloper/bp-go-lib/kafka"
	"github.com/PharbersDeveloper/bp-go-lib/log"
	"github.com/PharbersDeveloper/bp-jobs-observer/models"
	"github.com/PharbersDeveloper/bp-jobs-observer/models/record"
	"github.com/PharbersDeveloper/bp-jobs-observer/utils"
	"github.com/hashicorp/go-uuid"
	"os"
	"time"
)

func (observer *ObserverInfo) queryJobs() ([]record.HiveTracebackTask, error) {

	logger := log.NewLogicLoggerBuilder().Build()
	logger.Info("query jobs from db")

	newTraceId, err := uuid.GenerateUUID()

	var datasets []models.BpDataset

	//从 file 中取出 url array
	//TODO:check empty file
	file, err := os.Open(observer.FilePath)
	if err != nil {
		utils.Check(err)
	}

	scanner := bufio.NewScanner(file)
	var urlArr []interface{}
	for scanner.Scan() {
		urlArr = append(urlArr, scanner.Text())
	}
	utils.Check(scanner.Err())
	utils.Check(file.Close())

	//TODO: 抽离 magic word
	observer.Conditions["url"] = map[string]interface{}{"$in": urlArr}

	err = dbSession.DB(observer.Database).C(observer.Collection).Find(observer.Conditions).All(&datasets)
	if err != nil {
		return nil, err
	}

	expired := time.Duration(observer.SingleJobTimeoutSecond) * time.Second

	jobs := make([]record.HiveTracebackTask, 0)
	for _, dataset := range datasets {
		datasetId := dataset.Id.Hex()

		//TODO:在此处使用redis来check是否存在相同的job内容
		//TODO:暂时使用redis来存储以 checkKey 为job内容的唯一标示
		checkKey := "traceback_" + datasetId
		exist, e := utils.CheckKeyExistInRedis(checkKey)
		if e != nil {
			return nil, e
		}
		if !exist {
			count, e := utils.AddKey2Redis(checkKey)
			if e != nil {
				return nil, e
			}
			logger.Infof("将datasetId=%s的dataset加入job队列，redis inc count=%d", checkKey, count)

			ok, e := utils.SetKeyExpire(checkKey, expired)
			if !ok {
				logger.Infof("设置datasetId=%s过期时间失败，尝试重新设置过期时间", checkKey)
				_, _ = utils.SetKeyExpire(checkKey, expired)
			}

			//TODO:此处为拼接Job
			//TODO:此处使用UUID生成JobId
			newJobId, err := uuid.GenerateUUID()
			if err != nil {
				return nil, e
			}
			parentDatasetIds := make([]string, 0)
			parentUrl := new(record.ParentUrlRecord)
			{
				var newDatasets []models.BpDataset
				//TODO: 抽离 magic word
				newCondition := map[string]interface{}{"_id": map[string]interface{}{"$in": dataset.Parent}}
				err = dbSession.DB(observer.Database).C(observer.Collection).Find(newCondition).All(&newDatasets)
				if err != nil {
					return nil, err
				}
				for _, v := range newDatasets {
					parentDatasetIds = append(parentDatasetIds, v.Id.Hex())
					switch v.Description {
					case "MetaData":
						parentUrl.MetaData = v.Url
					case "SampleData":
						parentUrl.SampleData = v.Url
					default:
						logger.Infof("Unsupported  parent dataset description = %s", v.Description)
					}
				}
			}
			{
				job := record.HiveTracebackTask{
					JobId:           newJobId,
					TraceId:         newTraceId,
					DatasetId:       datasetId,
					ParentDatasetId: parentDatasetIds,
					ParentUrl:       parentUrl,
					Length:          dataset.Length,
					TaskType:        "append", //TODO：暂时只是创建增量的hive表的请求
					Remarks:         "",       //TODO：备注，预留字段
				}
				jobs = append(jobs, job)
			}

		}

	}

	return jobs, nil
}

func pushJobs(jobChan chan<- record.HiveTracebackTask, jobs []record.HiveTracebackTask) {
	logger := log.NewLogicLoggerBuilder().Build()
	logger.Info("push jobs to chan")
	for _, job := range jobs {
		jobChan <- job
	}
}

func (observer *ObserverInfo) worker(id int, jobChan <-chan record.HiveTracebackTask, ctx context.Context) {
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

func sendJobRequest(topic string, job record.HiveTracebackTask) error {

	specificRecordByteArr, err := kafka.EncodeAvroRecord(&job)
	if err != nil {
		return err
	}

	err = producer.Produce(topic, []byte(job.TraceId), specificRecordByteArr)
	return err
}

func (observer *ObserverInfo) scheduleJob(jobChan chan record.HiveTracebackTask, ctx context.Context) {

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