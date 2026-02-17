package core

import (
	"context"
)

type FetchResultInterface interface {
	Code() int64
	Err() error
	Source() string
	SetSource(string)
	SetError(error)
}

type fetchResult struct {
	code   int64
	err    error
	source string
	ctx    context.Context
}

func (r *fetchResult) Code() int64 {
	return r.code
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

func NewFetchResult(code int64) FetchResultInterface {
	return &fetchResult{
		code: code,
	}
}
