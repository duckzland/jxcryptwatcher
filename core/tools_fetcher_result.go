package core

import (
	"context"
)

type FetchResultInterface interface {
	Code() int64
	Data() any
	Err() error
	Source() string
	SetSource(string)
	SetError(error)
}

type fetchResult struct {
	code   int64
	data   any
	err    error
	source string
	ctx    context.Context
}

func (r *fetchResult) Code() int64 {
	return r.code
}

func (r *fetchResult) Data() any {
	return r.data
}

func (r *fetchResult) Err() error {
	return r.err
}

func (r *fetchResult) Source() string {
	return r.source
}

func (r *fetchResult) SetSource(s string) {
	r.source = s
}

func (r *fetchResult) SetError(e error) {
	r.err = e
}

func NewFetchResult(code int64, data any) FetchResultInterface {
	return &fetchResult{
		code: code,
		data: data,
	}
}
