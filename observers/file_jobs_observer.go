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
	KEY_JOBID = "JobId"
	KEY_ERROR = "Error"
)

var (
	dbSession    *mgo.Session
	kafkaBuilder *kafka.BpKafkaBuilder
	producer     *kafka.BpProducer
	consumer     *kafka.BpConsumer
	jobs         chan models.BpFile
	result       chan map[string]interface{}
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
	if err != nil {
		panic(err.Error())
	}

	length := len(assets)
	execLogger.Info("length=", length)

	//jobs = make(chan models.BpFile, bfjo.ParallelNumber)
	jobs = make(chan models.BpFile, length)
	result = make(chan map[string]interface{}, length)

	//分配worker执行Job
	for w := 1; w <= bfjo.ParallelNumber; w++ {
		go bfjo.worker(w, jobs)
	}

	//push job
	for _, a := range assets {
		file, err := bfjo.queryFile(a.File)
		if err != nil {
			panic(err)
		}
		jobs <- file
	}

	execLogger.Info("All jobs pushed done!")

	//启动监听返回值（阻塞当前线程）
	err = consumer.Consume(bfjo.ResponseTopic, subscribeAvroFunc)

	//bfjo.Close()
	execLogger.Info("End!")
}

func (bfjo *BpFileJobsObserver) Close() {
	close(jobs)
	close(result)
}

func (bfjo *BpFileJobsObserver) worker(id int, jobs <-chan models.BpFile) {
	workerLogger := log.NewLogicLoggerBuilder().Build()
	workerLogger.Info("worker", id, " standby!")

	for j := range jobs {

		jobLogger := log.NewLogicLoggerBuilder().SetJobId(j.Id.Hex()).Build()

		//send job request
		jobLogger.Info("worker", id, " started  job=", j.Id.Hex())
		err := sendJobRequest(bfjo.RequestTopic, j)
		if err != nil {
			panic(err)
		}

		//设置result chan的等待超时
		//t := time.NewTimer(6 * time.Hour) // 单个Job等待最高6小时
		t := time.NewTimer(6 * time.Second) // 测试使用：单个Job等待最高6s

		select {
		case resultMap := <- result:
			//处理返回值result
			bfjo.dealJobResult(resultMap)
			jobLogger.Info("worker", id, " finished job=", j.Id.Hex())
		case <-t.C:
			// 一直没有从result中读取到数据，Job超时，进行reDo，可配置超时策略？
			bfjo.reDoJob(j.Id.Hex())
			t.Stop()
			jobLogger.Info("worker", id, " run job=", j.Id.Hex(), " timeout.")
		}


	}
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
		JobId: job.Id.String(),
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
	//result <- err
	rMap := make(map[string]interface{}, 0)
	rMap[KEY_JOBID] = response.JobId
	rMap[KEY_ERROR] = response.Error
	result <- rMap
}

func (bfjo *BpFileJobsObserver) dealJobResult(resultMap map[string]interface{}) {

	var err string
	var jobId string

	e, ok := resultMap[KEY_ERROR]; if ok {
		err = e.(string)
	}
	j, ok := resultMap[KEY_JOBID]; if ok {
		jobId = j.(string)
	}

	if !(err == "" || err == "ok") {
		bfjo.reDoJob(jobId)
	}

}

func (bfjo *BpFileJobsObserver) reDoJob(jobId string)  {
	reDoLogger := log.NewLogicLoggerBuilder().SetJobId(jobId).Build()
	if bson.IsObjectIdHex(jobId) {
		id := bson.ObjectIdHex(jobId)
		file, err := bfjo.queryFile(id)
		if err != nil {
			reDoLogger.Error(err.Error())
		}
		jobs <- file
	} else {
		reDoLogger.Infof("Invalid ObjectId \"%s\"", jobId)
	}

}
