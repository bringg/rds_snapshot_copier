# RDS Snapshot copier

> **Deprication notice:** this tool is depricated, since AWS provides a built in [capability](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/USER_ReplicateBackups.html) to copy RDS backups.

This utility allows copying snapshots of AWS RDS instances.

## Features

- Copy snapshots between AWS regions
- Retention management of copied snapshot
- Automatically detects most recent snapshot
- Automatically generates target snapshot name based on the source ID

## Install

```shell
go get -v -u github.com/bringg/rds_snapshot_copier
```

## With Docker

```shell
docker run --rm bringg/rds-snapshot-copier
```

## Usage

```plain
Usage of rds_snapshot_copier:
    -copy-tags
            Copy all tags from the source snapshot to the target snapshot. (default true)
    -db-name string
            Source DB instance name.
    -kms-key-id string
            KMS key ID or ARN in target region, when specified the snapshot copy will be encrypted.
    -progress-timeout int
            Timeout in minutes when copy operation isn't progressing. (default 60)
    -retention int
            After successful copy, remove snapshots older than specified retention days. (default 30)
    -source-region string
            Region where the snapshot located.
    -target-region string
            Region where the snapshot will be copied to. (default same as source-region)
```

## Required IAM policy

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "rds:DescribeDBSnapshots",
                "rds:DeleteDBSnapshot",
                "rds:CopyDBSnapshot"
            ],
            "Resource": "*"
        }
    ]
}
```

## License

 Licensed under the MIT License. See the LICENSE file for details.
