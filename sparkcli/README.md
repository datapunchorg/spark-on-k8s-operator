# sparkcli

`sparkcli` is a command-line tool to submit Spark application to API Gateway in Spark Operator. It supports submitting, checking status, getting logs, and deleting Spark application.

## How to Build

Run command like following:

```
$ go build -o ~/sparkcli sparkcli/main.go
```

## Flags

The following are global flags for all sub commands in `sparkcli`:
* `--url`: the API Gateway url. Default value is `http://localhost:8080/sparkapi/v1`
* `--insecure`: skip SSL certificate verification if using a HTTPS API Gateway url
* `--user`: user name to connect to API gateway
* `--password`: user password to connect to API gateway

## Available Commands

### submit

`submit` is a sub command to submit Spark application to API Gateway in Spark Operator:

Usage Example:
```
sparkcli --user xxx --password xxx --insecure \
--url https://xxx.us-west-1.elb.amazonaws.com/sparkapi/v1 \
submit --image ghcr.io/datapunchorg/spark:pyspark-3.1-1643212945 --spark-version 3.1 \
--driver-memory 512m --executor-memory 512m \
your-pyspark-application.py
```

The `submit` command will get a submission id from API Gateway, and wait for the application to finish, like following:
```
2022/02/06 12:45:54 Submitted application, submission id: app-b36e008a380646ea8cf009919e9ef86d
2022/02/06 12:45:54 Waiting until application app-b36e008a380646ea8cf009919e9ef86d finished (current state: UNKNOWN)

2022/02/06 12:46:04 Waiting until application app-b36e008a380646ea8cf009919e9ef86d finished (current state: SUBMITTED)
2022/02/06 12:46:55 Waiting until application app-b36e008a380646ea8cf009919e9ef86d finished (current state: RUNNING)
...
2022/02/06 12:47:05 Waiting until application app-b36e008a380646ea8cf009919e9ef86d finished (current state: RUNNING)
2022/02/06 12:47:15 Application app-b36e008a380646ea8cf009919e9ef86d finished: {
    "submissionId": "app-b36e008a380646ea8cf009919e9ef86d",
    "state": "COMPLETED",
    "recentAppId": "spark-f7b574c02aa94e5ea5ec530691db6aec"
}
```

### status

`status` is a sub command to check Spark application status from API Gateway:

Usage Example:
```
sparkcli --user xxx --password xxx --insecure \
--url https://xxx.us-west-1.elb.amazonaws.com/sparkapi/v1 \
status app-b36e008a380646ea8cf009919e9ef86d
```

### log

`log` is a sub command to get Spark application log from API Gateway:

Usage Example:
```
sparkcli --user xxx --password xxx --insecure \
--url https://xxx.us-west-1.elb.amazonaws.com/sparkapi/v1 \
log app-b36e008a380646ea8cf009919e9ef86d
```
