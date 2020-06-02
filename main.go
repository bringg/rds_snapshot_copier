package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/rds"
)

func newAWSSession(region string) *session.Session {
	config := aws.NewConfig().WithCredentialsChainVerboseErrors(true).WithRegion(region)
	sess := session.Must(session.NewSession(config))

	if _, err := (sess.Config.Credentials.Get()); err != nil {
		log.Fatal(FormatAWSError(err, "newAWSSession"))
	}

	return sess
}

func copyDBInstance(targetRDS *rds.RDS, input *rds.CopyDBSnapshotInput, pTimeout time.Duration) error {
	output, err := targetRDS.CopyDBSnapshot(input)
	if err != nil {
		return err
	}

	describeInput := &rds.DescribeDBSnapshotsInput{DBSnapshotIdentifier: output.DBSnapshot.DBSnapshotIdentifier}
	progressed := time.Now()
	var progressPct int64

	log.Printf("copying snapshot to %s/%s ...", *input.DestinationRegion, *input.TargetDBSnapshotIdentifier)
	for range time.Tick(time.Second * 10) {
		if time.Since(progressed) >= pTimeout {
			return fmt.Errorf("snapshot copy isn't progressing, timed out reached after %s", pTimeout)
		}

		o, err := targetRDS.DescribeDBSnapshots(describeInput)
		if err != nil || len(o.DBSnapshots) != 1 {
			log.Print("failed to get copy progress...")
			continue
		}

		if *o.DBSnapshots[0].Status == "available" {
			log.Printf("copy completed!")
			break
		}

		if *o.DBSnapshots[0].PercentProgress > progressPct {
			progressed = time.Now()
		}

		progressPct = *o.DBSnapshots[0].PercentProgress
		log.Printf("%d%%, still copying... ", progressPct)
	}

	return nil
}

func main() {
	// standard logger: print filename and line number, without date/time
	log.SetFlags(log.Lshortfile)

	var (
		copyTags  = flag.Bool("copy-tags", true, "Copy all tags from the source snapshot to the target snapshot.")
		dbName    = flag.String("db-name", "", "Source DB instance name.")
		kmsKey    = flag.String("kms-key-id", "", "KMS key ID or ARN in target region, when specified the snapshot copy will be encrypted.")
		retention = flag.Int("retention", 30, "After successful copy, remove snapshots older than specified retention days.")
		sRegion   = flag.String("source-region", "", "Region where the snapshot located.")
		timeout   = flag.Int("progress-timeout", 60, "Timeout in minutes when copy operation isn't progressing.")
		tRegion   = flag.String("target-region", "", "Region where the snapshot will be copied to. (default same as source-region)")
	)
	flag.Parse()

	if *dbName == "" {
		log.Fatal("db-name is required argument.")
	}

	if *tRegion == "" {
		tRegion = sRegion
	}

	sourceRDS := rds.New(newAWSSession(*sRegion))
	sourceDB := MustDBInstance(NewDBInstance(*dbName, sourceRDS))

	log.Printf("getting most recent snapshot of %s/%s RDS instance", *sRegion, *dbName)
	lastSnapshot, err := sourceDB.GetLastSnapshot()
	if err != nil {
		log.Fatal(FormatAWSError(err, "sourceDB.GetLastSnapshot"))
	}

	log.Printf("found recent snapshot %s", *lastSnapshot.DBSnapshotIdentifier)

	// prepare target
	targetRDS := rds.New(newAWSSession(*tRegion))
	targetDB := MustDBInstance(NewDBInstance(*dbName, targetRDS))
	targetDBName := aws.String(fmt.Sprintf("%s-%s", *dbName, lastSnapshot.SnapshotCreateTime.Format("2006-01-02-15-04")))
	input := &rds.CopyDBSnapshotInput{
		CopyTags:                   copyTags,
		SourceRegion:               sRegion,
		SourceDBSnapshotIdentifier: lastSnapshot.DBSnapshotArn,
		TargetDBSnapshotIdentifier: targetDBName,
	}

	if *kmsKey != "" {
		log.Printf("using [%s] KMS key ID for encryption", *kmsKey)
		input.KmsKeyId = kmsKey
	}

	if err := copyDBInstance(targetRDS, input, time.Minute*time.Duration(*timeout)); err != nil {
		log.Fatal(FormatAWSError(err, "copyDBInstance"))
	}

	// delete snapshots on target DB older than specified days
	log.Printf("looking for old snapshots which match %d retention days...", *retention)
	oldSnapshots := targetDB.GetOldSnapshots(*retention)

	log.Printf("found %d snapshots to delete", len(oldSnapshots))
	for _, s := range oldSnapshots {
		log.Print("deleting snapshot:", *s.DBSnapshotIdentifier)
		if _, err := targetRDS.DeleteDBSnapshot(&rds.DeleteDBSnapshotInput{DBSnapshotIdentifier: s.DBSnapshotIdentifier}); err != nil {
			log.Print(FormatAWSError(err, "targetRDS.DeleteDBSnapshot"))
		}
	}

	log.Print("all done!")
}
