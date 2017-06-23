package main

import (
	"fmt"
	"sort"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
)

// DBInstance ...
type DBInstance struct {
	ID        string
	RDS       *rds.RDS
	snapshots []*rds.DBSnapshot
}

// NewDBInstance returns an initialized DBInstance instance
func NewDBInstance(id string, rds *rds.RDS) (*DBInstance, error) {
	dbInstance := &DBInstance{
		ID:  id,
		RDS: rds,
	}

	if err := dbInstance.GetSnapshots(); err != nil {
		return nil, err
	}

	return dbInstance, nil
}

// MustDBInstance is a helper function to ensure the *DBInstance is valid
// and there was no error when calling a NewDBInstance function.
func MustDBInstance(instance *DBInstance, err error) *DBInstance {
	if err != nil {
		panic(err)
	}

	return instance
}

// GetSnapshots gets a slice of all snapshots of the *DBinstance,
// sorted by Creation time
func (i *DBInstance) GetSnapshots() error {
	output, err := i.RDS.DescribeDBSnapshots(&rds.DescribeDBSnapshotsInput{
		DBInstanceIdentifier: aws.String(i.ID),
	})

	if err != nil {
		return err
	}

	// filter non-available snapshots from the DBSnapshots slice
	// no allocations used, see: https://github.com/golang/go/wiki/SliceTricks
	filtered := output.DBSnapshots[:0]
	for _, s := range output.DBSnapshots {
		if *s.Status == "available" {
			filtered = append(filtered, s)
		}
	}

	// sort by snapshot creation time
	sort.Slice(filtered, func(i, j int) bool {
		return (*filtered[i].SnapshotCreateTime).Before(*filtered[j].SnapshotCreateTime)
	})

	i.snapshots = filtered
	return nil
}

// GetLastSnapshot returns a pointer to most recent Snapshot
func (i DBInstance) GetLastSnapshot() (*rds.DBSnapshot, error) {
	snapshots := i.snapshots

	if len(snapshots) > 0 {
		return snapshots[len(snapshots)-1], nil
	}

	return nil, fmt.Errorf("couldn't get last snapshot for %s instance, no available snapshots found", i.ID)
}

// GetOldSnapshots returns a slice of pointers to all snapshots which are older
// than specified "days"
func (i DBInstance) GetOldSnapshots(days int) ([]*rds.DBSnapshot, error) {
	var oldSnapshots []*rds.DBSnapshot
	oldDate := time.Now().AddDate(0, 0, -days)

	for _, s := range i.snapshots {
		if *s.Status != "available" {
			continue
		}

		if s.SnapshotCreateTime.After(oldDate) {
			// the i.Snapshots slice is sorted,
			// so we can safely break here
			break
		}

		oldSnapshots = append(oldSnapshots, s)
	}

	return oldSnapshots, nil
}
