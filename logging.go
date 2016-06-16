package main

import "github.com/kovetskiy/lorg"

const (
	logFormatString = `${level} %s`
)

var (
	logger = lorg.NewLog()
)

func init() {
	format := lorg.NewFormat(logFormatString)
	logger.SetFormat(format)
}
