package factory

import (
	"errors"
	"reflect"
	"sync"
)

type bpFactory struct {
	m map[string]reflect.Type
}

var fac *bpFactory
var once sync.Once

func NewFactory() *bpFactory {
	once.Do(func() {
		fac = &bpFactory{
			m:make(map[string]reflect.Type),
		}
	})
	return fac
}

func (fac *bpFactory) RegisterModel(name string, tp interface{}) {
	for k, _ := range fac.m {
		if k == name {
			return
		}
	}
	t := reflect.TypeOf(tp).Elem()
	fac.m[name] = t
}

func (fac *bpFactory) ReflectInstance(name string) (interface{}, error) {
	if typ, ok := fac.m[name]; ok {
		reva := reflect.New(typ).Elem().Interface()
		return reva, nil
	} else {
		return nil, errors.New("not register model")
	}
}
