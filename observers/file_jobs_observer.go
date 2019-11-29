package observers

import (
	"fmt"
	"github.com/PharbersDeveloper/bp-go-lib/kafka"
	"github.com/PharbersDeveloper/bp-go-lib/kafka/record"
	"github.com/PharbersDeveloper/bp-go-lib/log"
	"github.com/PharbersDeveloper/bp-jobs-observer/models"
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
	//jobs         chan models.BpFile
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
		panic(err)
	}
	sess.SetMode(mgo.Monotonic, true)

	dbSession = sess

	kafkaBuilder = kafka.NewKafkaBuilder()
	producer, err = kafkaBuilder.BuildProducer()
	consumer, err = kafkaBuilder.SetGroupId(bfjo.Id).BuildConsumer()
	if err != nil {
		panic(err.Error())
	}

}

func (bfjo *BpFileJobsObserver) Exec() {
	execLogger := log.NewLogicLoggerBuilder().Build()
	execLogger.Info("start exec")

	//查询Jobs
	files, err := bfjo.queryJobs()
	if err != nil {
		panic(err.Error())
	}

	length := len(files)
	execLogger.Info("jobs length=", length)

	jobs := make(chan models.BpFile, length)
	//用作判断job是否完成的
	jobStatus = make(chan map[string]string, 1)
	jStatus := make(map[string]string, 0)
	//初始化，存进jobStatus
	jobStatus <- jStatus

	//分配worker执行Job
	for id := 1; id <= bfjo.ParallelNumber; id++ {
		go bfjo.worker(id, jobs)
	}

	//将file job push 到 chan队列
	pushJobs(jobs, files)
	execLogger.Info("All jobs pushed done!")

	//另起一个协程执行scheduleJob，定时刷新job队列
	go bfjo.scheduleJob(jobs)

	//启动监听返回值（阻塞当前线程）
	err = consumer.Consume(bfjo.ResponseTopic, subscribeAvroFunc)

	//bfjo.Close()
	execLogger.Info("End!")
}

func (bfjo *BpFileJobsObserver) Close() {
	//close(jobs)
	close(jobStatus)
}

func (bfjo *BpFileJobsObserver) queryJobs() ([]models.BpFile, error) {

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

	files := make([]models.BpFile, 0)
	for _, a := range assets {
		file, e := bfjo.queryFile(a.File)
		if e != nil {
			return nil, e
		}
		files = append(files, file)
	}

	return files, nil
}

func (bfjo *BpFileJobsObserver) queryFile(id bson.ObjectId) (models.BpFile, error) {
	var file models.BpFile
	err := dbSession.DB(bfjo.Database).C("files").Find(bson.M{"_id": id}).One(&file)
	if err != nil {
		return file, err
	}
	return file, err
}

func pushJobs(jobs chan<- models.BpFile, files []models.BpFile) {
	logger := log.NewLogicLoggerBuilder().Build()
	logger.Info("push jobs to chan")
	for _, file := range files {
		jobs <- file
	}
}

func (bfjo *BpFileJobsObserver) worker(id int, jobs <-chan models.BpFile) {
	workerLogger := log.NewLogicLoggerBuilder().Build()
	workerLogger.Info("worker", id, " standby!")

	//设置等待超时
	d := time.Duration(bfjo.SingleJobTimeoutSecond) * time.Second
	t := time.NewTimer(d) // 测试使用：单个Job等待时长

	for j := range jobs {

		jobId := j.Id.Hex()
		jobLogger := log.NewLogicLoggerBuilder().SetJobId(jobId).Build()

		//send job request
		jobLogger.Info("worker", id, " started  job=", jobId)
		err := sendJobRequest(bfjo.RequestTopic, j)
		if err != nil {
			panic(err)
		}

		//取出jobStatus
		jStatus := <-jobStatus
		jStatus[jobId] = JOB_RUN
		//存进jobStatus
		jobStatus <- jStatus

		locked := true
		for locked {

			select {
			case <-t.C:
				//超时后不进行reDo，reDo会造成重复计算，设置超时策略
				jobLogger.Info("worker", id, " run job=", jobId, " timeout.")
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
			}

		}
		t.Reset(d)

	}

	workerLogger.Info("worker", id, " done!")
}

func sendJobRequest(topic string, job models.BpFile) error {

	requestRecord := record.ExampleRequest{
		JobId: job.Id.Hex(),
		Tag:   job.Extension,
		Configs: []string{
			job.FileName,
			job.Url,
		},
	}

	specificRecordByteArr, err := kafka.EncodeAvroRecord(&requestRecord)
	if err != nil {
		return err
	}

	err = producer.Produce(topic, []byte("avro-key003"), specificRecordByteArr)
	return err
}

func subscribeAvroFunc(key interface{}, value interface{}) {

	var response record.ExampleResponse
	//注意传参为record的地址
	err := kafka.DecodeAvroRecord(value.([]byte), &response)
	if err != nil {
		panic(err.Error())
	}
	logger := log.NewLogicLoggerBuilder().SetJobId(response.JobId).Build()
	logger.Infof("response => key=%s, value=%v\n", string(key.([]byte)), response)
	jobId := response.JobId
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

	//logger := log.NewLogicLoggerBuilder().SetJobId(jobId).Build()

	status, ok := jStatus[jobId]

	if ok {
		switch status {
		case JOB_RUN:
		case JOB_END:
			//logger.Infof("job=%s is done.", jobId)
			done = true
		case JOB_ERROR:
		}

	}

	return

}

func (bfjo *BpFileJobsObserver) scheduleJob(jobs chan models.BpFile) {

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
		case <-t.C:
			logger.Info("start schedule job ...")

			scanCurrentJobs(jobs)

			//查询Jobs
			files, err := bfjo.queryJobs()
			if err != nil {
				panic(err.Error())
			}

			length := len(files)
			logger.Info("jobs length=", length)

			pushJobs(jobs, files)

			// need reset timer
			t.Reset(d)
		}

	}

}

func scanCurrentJobs(jobs chan models.BpFile) {

	logger := log.NewLogicLoggerBuilder().Build()
	logger.Info("Start scan current jobs")
	jobsChanNotEmpty := len(jobs) != 0
	jobsIsRunning := true
	//1. 每隔1s查询一次jobs chan 是否为空
	for jobsChanNotEmpty {
		time.Sleep(time.Second)
		jobsChanNotEmpty = len(jobs) != 0
	}
	//2. 每隔1s查询JobStatus的状态
	for jobsIsRunning {
		time.Sleep(time.Second)
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
