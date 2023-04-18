### Localstack Issue

This is a small demo app to demonstrate the problem I'm seeing.

In `docker-compose.yml` you can see that we specify two services:

* `endpoint` the web server
* `localstack` the local version of DynamoDB


**List Dynamo tables**: `aws --endpoint-url http://localhost:4566 dynamodb list-tables`

**NoSQL Workbench:**
https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/workbench.html