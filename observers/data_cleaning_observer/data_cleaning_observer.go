package data_cleaning_observer

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

func (observer *ObserverInfo) Open()  {

}

func (observer *ObserverInfo) Exec()  {

}

func (observer *ObserverInfo) Close() {

}
