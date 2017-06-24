## RDS Snapshot copier

This utility allows copying snapshots of AWS RDS instances.

**Features**

- Copy snapshots between AWS regions
- Retention management of copied snapshot
- Automatically detects most recent snapshot
- Automatically generates target snapshot name based on the source ID

**Install**

    go get -v -u  github.com/bringg/rds_snapshot_copier

**With Docker**

    docker run --rm bringg/rds-snapshot-copier

**Usage**

    Usage of rds_snapshot_copier:
    -db-name string
            Source DB instance name.
    -kms-key-id string
            KMS key ID or ARN in target region, when specified the snapshot copy will be encrypted.
    -retention int
            After successful copy, remove snapshots older than specified retention days. (default 30)
    -source-region string
            Region where the snapshot located.
    -target-region string
            Region where the snapshot will be copied to. (default same as source-region)
    -timeout int
            Number of minutes to wait for copy operation completion (default 60)


 **License**

 Licensed under the MIT License. See the LICENSE file for details.
