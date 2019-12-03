package observers

import (
	"context"
	"fmt"
	"github.com/PharbersDeveloper/bp-go-lib/kafka"
	"github.com/PharbersDeveloper/bp-go-lib/log"
	"github.com/PharbersDeveloper/bp-jobs-observer/models"
	"github.com/PharbersDeveloper/bp-jobs-observer/models/record"
	"github.com/hashicorp/go-uuid"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"time"
)

type BpFileJobsObserver struct {
	Id                     string                 `json:"id"`
	DBHost                 string                 `json:"db_host"`
	DBPort                 string                 `json:"db_port"`
	Database               string                 `json:"database"`
	Collection             string                 `json:"collection"`
	Conditions             map[string]interface{} `json:"conditions"`
	ParallelNumber         int                    `json:"parallel_number"`
	SingleJobTimeoutSecond int64                    `json:"single_job_timeout_second"`
	ScheduleDurationSecond int64                    `json:"schedule_duration_second"`
	RequestTopic           string                 `json:"request_topic"`
	ResponseTopic          string                 `json:"response_topic"`
}

const (
	JOB_RUN   = "run"
	JOB_END   = "end"
	JOB_ERROR = "error"
)

var (
	dbSession    *mgo.Session
	kafkaBuilder *kafka.BpKafkaBuilder
	producer     *kafka.BpProducer
	consumer     *kafka.BpConsumer
	jobStatus    chan map[string]string
)

func (bfjo *BpFileJobsObserver) Open() {
	mongoDBDialInfo := &mgo.DialInfo{
		Addrs:   []string{fmt.Sprintf("%s:%s", bfjo.DBHost, bfjo.DBPort)},
		Timeout: 1 * time.Hour,
	}

	sess, err := mgo.DialWithInfo(mongoDBDialInfo)
	if err != nil {
		sess.Refresh()
		log.NewLogicLoggerBuilder().Build().Error(err.Error())
	}
	sess.SetMode(mgo.Monotonic, true)

	dbSession = sess

	kafkaBuilder = kafka.NewKafkaBuilder()
	producer, err = kafkaBuilder.BuildProducer()
	consumer, err = kafkaBuilder.SetGroupId(bfjo.Id).BuildConsumer()
	if err != nil {
		log.NewLogicLoggerBuilder().Build().Error(err.Error())
	}

}

func (bfjo *BpFileJobsObserver) Exec() {
	execLogger := log.NewLogicLoggerBuilder().Build()
	execLogger.Info("start exec")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//查询Jobs
	jobs, err := bfjo.queryJobs()
	if err != nil {
		execLogger.Error(err.Error())
	}

	length := len(jobs)
	execLogger.Info("jobs length=", length)

	jobChan := make(chan record.OssTask, length)
	defer close(jobChan)

	//用作判断job是否完成的
	jobStatus = make(chan map[string]string, 1)
	defer  close(jobStatus)

	//初始化，存进jobStatus
	jStatus := make(map[string]string, 0)
	jobStatus <- jStatus

	//分配worker执行Job
	for id := 1; id <= bfjo.ParallelNumber; id++ {
		go bfjo.worker(id, jobChan, ctx)
	}

	//将file job push 到 chan队列
	pushJobs(jobChan, jobs)
	execLogger.Info("All jobChan pushed done!")

	//另起一个协程执行scheduleJob，定时刷新job队列
	go bfjo.scheduleJob(jobChan, ctx)

	//启动监听返回值（阻塞当前线程）
	err = consumer.Consume(bfjo.ResponseTopic, subscribeAvroFunc)
	//select {
	//case <-ctx.Done():
	//	execLogger.Info("Exec context done!")
	//}

	//bfjo.Close()
	execLogger.Info("End!")
}

func (bfjo *BpFileJobsObserver) Close() {
	//close(jobChan)
	//close(jobStatus)
}

func (bfjo *BpFileJobsObserver) queryJobs() ([]record.OssTask, error) {

	logger := log.NewLogicLoggerBuilder().Build()
	logger.Info("query jobs from db")
	var assets []models.BpAsset
	//err := dbSession.DB(bfjo.Database).C(bfjo.Collection).Find(bfjo.Conditions).All(&assets)
	//临时测试12个
	err := dbSession.DB(bfjo.Database).C(bfjo.Collection).Find(bfjo.Conditions).Limit(5).All(&assets)
	if err != nil {
		return nil, err
	}

	count, err := dbSession.DB(bfjo.Database).C(bfjo.Collection).Find(bfjo.Conditions).Count()
	if err != nil {
		return nil, err
	}
	logger.Info("count=", count)

	jobs := make([]record.OssTask, 0)
	for _, asset := range assets {
		file, e := bfjo.queryFile(asset.File)
		if e != nil {
			return nil, e
		}
		job := record.OssTask{
			TitleIndex: nil,
			JobId:      "",
			TraceId:    asset.TraceId,
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

	return jobs, nil
}

func (bfjo *BpFileJobsObserver) queryFile(id bson.ObjectId) (models.BpFile, error) {
	var file models.BpFile
	err := dbSession.DB(bfjo.Database).C("files").Find(bson.M{"_id": id}).One(&file)
	if err != nil {
		return file, err
	}
	return file, err
}

func pushJobs(jobChan chan<- record.OssTask, jobs []record.OssTask) {
	logger := log.NewLogicLoggerBuilder().Build()
	logger.Info("push jobs to chan")
	for _, file := range jobs {
		jobChan <- file
	}
}

func (bfjo *BpFileJobsObserver) worker(id int, jobChan <-chan record.OssTask, ctx context.Context) {
	workerLogger := log.NewLogicLoggerBuilder().Build()
	workerLogger.Info("worker", id, " standby!")

	//设置等待超时
	d := time.Duration(bfjo.SingleJobTimeoutSecond) * time.Second
	t := time.NewTimer(d) // 测试使用：单个Job等待时长

	for j := range jobChan {

		//TODO: 由于asset是以traceId为区分，jobId未使用，这里自动化Job以traceId作为JobId
		jobId := j.TraceId
		traceId := j.TraceId

		//压测使用
		{
			newTraceId, err := bfjo.updateTraceId(traceId)
			if err != nil {
				workerLogger.Error(err.Error())
			}
			jobId = newTraceId
			traceId = newTraceId
			j.JobId = newTraceId
			j.TraceId = newTraceId
		}

		jobLogger := log.NewLogicLoggerBuilder().SetTraceId(traceId).SetJobId(jobId).Build()

		//send job request
		jobLogger.Infof("worker-%d started  job=%v", id,  j)
		err := sendJobRequest(bfjo.RequestTopic, j)
		if err != nil {
			jobLogger.Error(err.Error())
		}

		//取出jobStatus
		jStatus := <-jobStatus
		jStatus[jobId] = JOB_RUN
		//存进jobStatus
		jobStatus <- jStatus

		locked := true
		for locked {

			select {
			case <-ctx.Done():
				//上层（调用协程的）结束，终止子协程
				jobLogger.Info("worker", id, " stop job=", jobId, ", because context is done.")
				return
			case <-t.C:
				//超时后不进行reDo，reDo会造成重复计算，设置超时策略
				jobLogger.Info("worker", id, " run job=", jobId, " timeout.")
				//暂时超时后，就不管这个jobId了
				//取出jobStatus
				jStatus := <-jobStatus
				delete(jStatus, jobId)
				//存进jobStatus
				jobStatus <- jStatus
				locked = false
			//取出jobStatus
			case jStatus := <-jobStatus:
				//根据job status执行不同case
				done := bfjo.dealJobResult(jStatus, jobId)
				if done {
					locked = false
					jobLogger.Info("worker", id, " finished job=", jobId)
				}
				//存进jobStatus
				jobStatus <- jStatus
				time.Sleep(time.Second)
			}

		}
		t.Reset(d)

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

func subscribeAvroFunc(key interface{}, value interface{}) {

	var response record.OssTaskResult
	//注意传参为record的地址
	err := kafka.DecodeAvroRecord(value.([]byte), &response)
	if err != nil {
		log.NewLogicLoggerBuilder().Build().Error(err.Error())
	}

	//TODO: 由于asset是以traceId为区分，jobId未使用，这里自动化Job以traceId作为JobId
	jobId := response.TraceId
	traceId := response.TraceId
	logger := log.NewLogicLoggerBuilder().SetTraceId(traceId).SetJobId(jobId).Build()
	logger.Infof("response => key=%s, value=%v\n", string(key.([]byte)), response)

	//取出jobStatus
	jStatus := <-jobStatus
	if jStatus[jobId] == JOB_RUN {
		jStatus[jobId] = JOB_END
	}
	if response.Error != "" {
		jStatus[jobId] = JOB_ERROR
	}
	logger.Infof("response => jStatus=%v\n", jStatus)
	//存进jobStatus
	jobStatus <- jStatus
}

func (bfjo *BpFileJobsObserver) dealJobResult(jStatus map[string]string, jobId string) (done bool) {

	logger := log.NewLogicLoggerBuilder().SetJobId(jobId).Build()

	status, ok := jStatus[jobId]

	if ok {
		switch status {
		case JOB_RUN:
		case JOB_END:
			logger.Infof("job=%s is done.", jobId)
			done = true
		case JOB_ERROR:
		}

	}

	return

}

func (bfjo *BpFileJobsObserver) scheduleJob(jobChan chan record.OssTask, ctx context.Context) {

	//1. 设置ticker 循环时间
	//2. 扫描当前的Jobs，确定job len为空，且jobStatus都为End
	//3. 重新查询Jobs
	//4. 将查询的job push到job chan

	logger := log.NewLogicLoggerBuilder().Build()

	//设置时间周期
	d := time.Duration(bfjo.ScheduleDurationSecond) * time.Second
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

			scanCurrentJobs(jobChan)

			//压测使用，更新traceId，将dfs置空
			bfjo.restartJobChan()

			//查询Jobs
			files, err := bfjo.queryJobs()
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

func scanCurrentJobs(jobChan chan record.OssTask) {

	logger := log.NewLogicLoggerBuilder().Build()
	logger.Info("Start scan current jobs")
	jobsChanNotEmpty := len(jobChan) != 0
	jobsIsRunning := true
	//1. 每隔1min查询一次jobs chan 是否为空
	for jobsChanNotEmpty {
		time.Sleep(time.Minute)
		jobsChanNotEmpty = len(jobChan) != 0
	}
	//2. 每隔1min查询JobStatus的状态
	for jobsIsRunning {
		time.Sleep(time.Minute)
		jobsIsRunning = checkJobsIsRunning()
	}
	logger.Info("Current jobs all done")

}

func checkJobsIsRunning() bool {

	//取出jobStatus
	jStatus := <-jobStatus
	for _, status := range jStatus {
		if status == JOB_RUN {
			//存进jobStatus
			jobStatus <- jStatus
			return true
		}
	}
	//存进jobStatus
	jobStatus <- jStatus
	return false
}

//压测使用
func (bfjo *BpFileJobsObserver) restartJobChan() {

	//取出jobStatus
	jStatus := <-jobStatus
	for jobId, _ := range jStatus {
		bfjo.updateDfs(jobId)
	}
	//重新初始化，存进jobStatus
	jStatus = make(map[string]string, 0)
	jobStatus <- jStatus

}

//压测使用
func (bfjo *BpFileJobsObserver) updateTraceId(jobId string) (newJobId string, err error) {

	//TODO: 由于asset是以traceId为区分，jobId未使用，这里自动化Job以traceId作为JobId
	traceId := jobId
	var asset models.BpAsset
	err = dbSession.DB(bfjo.Database).C(bfjo.Collection).Find(bson.M{"traceId": traceId}).One(&asset)
	if err != nil {
		return "", err
	}

	newTraceId, err := uuid.GenerateUUID()
	asset.TraceId = newTraceId

	err = dbSession.DB(bfjo.Database).C(bfjo.Collection).UpdateId(asset.Id, asset)
	return newTraceId, err
}

//压测使用
func (bfjo *BpFileJobsObserver) updateDfs(jobId string) {

	//TODO: 由于asset是以traceId为区分，jobId未使用，这里自动化Job以traceId作为JobId
	traceId := jobId
	logger := log.NewLogicLoggerBuilder().SetJobId(jobId).SetTraceId(traceId).Build()
	var asset models.BpAsset
	err := dbSession.DB(bfjo.Database).C(bfjo.Collection).Find(bson.M{"traceId": traceId}).One(&asset)
	if err != nil {
		logger.Error(err.Error())
	}

	asset.Dfs = nil

	err = dbSession.DB(bfjo.Database).C(bfjo.Collection).UpdateId(asset.Id, asset)
	if err != nil {
		logger.Error(err.Error())
	}
}
