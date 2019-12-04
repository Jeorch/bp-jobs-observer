package oss_task_observer

import (
	"context"
	"fmt"
	"github.com/PharbersDeveloper/bp-go-lib/kafka"
	"github.com/PharbersDeveloper/bp-go-lib/log"
	"github.com/PharbersDeveloper/bp-jobs-observer/models/record"
	"gopkg.in/mgo.v2"
	"time"
)

type ObserverInfo struct {
	Id                     string                 `json:"id"`
	DBHost                 string                 `json:"db_host"`
	DBPort                 string                 `json:"db_port"`
	Database               string                 `json:"database"`
	Collection             string                 `json:"collection"`
	Conditions             map[string]interface{} `json:"conditions"`
	ParallelNumber         int                    `json:"parallel_number"`
	SingleJobTimeoutSecond int64                  `json:"single_job_timeout_second"`
	ScheduleDurationSecond int64                  `json:"schedule_duration_second"`
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
)

func (observer *ObserverInfo) Open() {
	mongoDBDialInfo := &mgo.DialInfo{
		Addrs:   []string{fmt.Sprintf("%s:%s", observer.DBHost, observer.DBPort)},
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
	//consumer, err = kafkaBuilder.SetGroupId(observer.Id).BuildConsumer()
	if err != nil {
		log.NewLogicLoggerBuilder().Build().Error(err.Error())
	}

}

func (observer *ObserverInfo) Exec() {
	execLogger := log.NewLogicLoggerBuilder().Build()
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

	jobChan := make(chan record.OssTask, length)
	defer close(jobChan)

	//分配worker执行Job
	for id := 1; id <= observer.ParallelNumber; id++ {
		go observer.worker(id, jobChan, ctx)
	}

	//将file job push 到 chan队列
	pushJobs(jobChan, jobs)
	execLogger.Info("All jobChan pushed done!")

	//另起一个协程执行scheduleJob，定时刷新job队列
	go observer.scheduleJob(jobChan, ctx)

	select {
	case <-ctx.Done():
		execLogger.Info("Exec context done!")
	}

	//observer.Close()
	execLogger.Info("End!")
}

func (observer *ObserverInfo) Close() {
	//close(jobChan)
	//close(jobStatus)
}
