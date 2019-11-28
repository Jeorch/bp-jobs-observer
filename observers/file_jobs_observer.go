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

//"args": {
//    "db_host": "59.110.31.50",
//    "db_port": "5555",
//    "database": "pharbers-sandbox-600",
//    "collection": "assets",
//    "conditions": [
//      {
//        "key": "file",
//        "value": "null",
//        "operator": "!="
//      },
//      {
//        "key": "dfs",
//        "value": "nil",
//        "operator": "=="
//      }
//    ]
//  }
type BpFileJobsObserver struct {
	Id             string                 `json:"id"`
	DBHost         string                 `json:"db_host"`
	DBPort         string                 `json:"db_port"`
	Database       string                 `json:"database"`
	Collection     string                 `json:"collection"`
	Conditions     map[string]interface{} `json:"conditions"`
	ParallelNumber int                    `json:"parallel_number"`
	RequestTopic   string                 `json:"request_topic"`
	ResponseTopic  string                 `json:"response_topic"`
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
	jobs         chan models.BpFile
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
	var assets []models.BpAsset
	//err := dbSession.DB(bfjo.Database).C(bfjo.Collection).Find(bfjo.Conditions).All(&assets)
	//临时测试12个
	err := dbSession.DB(bfjo.Database).C(bfjo.Collection).Find(bfjo.Conditions).Limit(12).All(&assets)
	count, err := dbSession.DB(bfjo.Database).C(bfjo.Collection).Find(bfjo.Conditions).Count()
	execLogger.Info("count=", count)
	if err != nil {
		panic(err.Error())
	}

	length := len(assets)
	execLogger.Info("length=", length)

	jobs = make(chan models.BpFile, length)
	//用作判断job是否完成的
	jobStatus = make(chan map[string]string, 1)
	jStatus := make(map[string]string, 0)
	//初始化，存进jobStatus
	jobStatus <- jStatus

	//分配worker执行Job
	for w := 1; w <= bfjo.ParallelNumber; w++ {
		go bfjo.worker(w, jobs)
	}

	//push job
	for i, a := range assets {
		file, err := bfjo.queryFile(a.File)
		if err != nil {
			panic(err)
		}
		fmt.Println(i, "push")
		jobs <- file
		fmt.Println(i, "push done")
	}

	execLogger.Info("All jobs pushed done!")

	//启动监听返回值（阻塞当前线程）
	err = consumer.Consume(bfjo.ResponseTopic, subscribeAvroFunc)

	//bfjo.Close()
	execLogger.Info("End!")
}

func (bfjo *BpFileJobsObserver) Close() {
	close(jobs)
	close(jobStatus)
}

func (bfjo *BpFileJobsObserver) worker(id int, jobs <-chan models.BpFile) {
	workerLogger := log.NewLogicLoggerBuilder().Build()
	workerLogger.Info("worker", id, " standby!")

	for j := range jobs {

		//取出jobStatus
		jStatus := <- jobStatus
		jobId := j.Id.Hex()
		jobLogger := log.NewLogicLoggerBuilder().SetJobId(jobId).Build()

		//send job request
		jobLogger.Info("worker", id, " started  job=", jobId)
		err := sendJobRequest(bfjo.RequestTopic, j)
		if err != nil {
			panic(err)
		}
		jStatus[jobId] = JOB_RUN
		//存进jobStatus
		jobStatus <- jStatus

		//设置等待超时
		//t := time.NewTimer(6 * time.Hour) // 单个Job等待最高6小时
		t := time.NewTimer(6 * time.Second) // 测试使用：单个Job等待最高6s

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
		t.Stop()

	}

	workerLogger.Info("worker", id, " done!")
}

func (bfjo *BpFileJobsObserver) queryFile(id bson.ObjectId) (models.BpFile, error) {
	var file models.BpFile
	err := dbSession.DB(bfjo.Database).C("files").Find(bson.M{"_id": id}).One(&file)
	if err != nil {
		panic(err.Error())
	}
	return file, err
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
	jStatus := <- jobStatus
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
