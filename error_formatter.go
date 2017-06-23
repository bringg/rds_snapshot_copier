package main

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws/awserr"
)

// FormatAWSError returns an instance of AwsErrorFormatter
// mostly used by functions that expect the Stringer interface
func FormatAWSError(err error, meta string) *AwsErrorFormatter {
	return &AwsErrorFormatter{
		Error: err,
		Meta:  meta,
	}
}

// AwsErrorFormatter is a primitive formatter of AWS errors
type AwsErrorFormatter struct {
	Error error
	Meta  string
}

// String implements the Stringer interface
func (f AwsErrorFormatter) String() string {
	err := f.Error

	if awsErr, ok := err.(awserr.Error); ok {
		return fmt.Sprintf("%s: %s", f.Meta, awsErr.Message())
	}

	return fmt.Sprintf("%s: %v", f.Meta, err)
}
