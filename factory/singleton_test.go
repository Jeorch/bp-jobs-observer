package factory

import (
	"fmt"
	"github.com/PharbersDeveloper/bp-jobs-observer/observers/observer_data_cleaning"
	"github.com/PharbersDeveloper/bp-jobs-observer/observers/observer_oss_task"
	"testing"
)

func TestBpFactory(t *testing.T) {
	fac := NewFactory()
	fac.RegisterModel("observer_oss_task", &observer_oss_task.ObserverInfo{})
	fac.RegisterModel("observer_data_cleaning", &observer_data_cleaning.ObserverInfo{})

	fmt.Println(fac.m)

	model, err := fac.ReflectInstance("observer_data_cleaning")
	if err != nil {
		panic(err.Error())
	}
	fmt.Println(model)

}
