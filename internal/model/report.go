package model

import (
	"time"
)

type Report struct {
	URL             string
	TotalTestCount  int64
	FinishTestCount int64
	TestResults     []TestResult
}

type TestResult struct {
	Type    string
	Results []Result
}

type Result struct {
	URL       string
	PoCs      []PoC
	StartTime time.Time
	EndTime   time.Time
}

type PoC struct {
	Type       string
	InjectType string
	PoCType    string
	Method     string
	Data       string
	Param      string
	Payload    string
	Evidence   string
	CWE        string
	Severity   string
}
