# Tyk Pump

[![Build Status](https://travis-ci.org/TykTechnologies/tyk-pump.svg?branch=master)](https://travis-ci.org/TykTechnologies/tyk-pump)

Tyk Pump is a pluggable analytics purger to move Analytics generated by your Tyk nodes to any back-end.

## Back ends currently supported:

- MongoDB (to replace built-in purging)
- CSV (updated, now supports all fields)
- ElasticSearch (2.0+)
- Graylog
- InfluxDB
- Moesif
- Splunk
- StatsD
- DogStatsD
- Hybrid (Tyk RPC)
- Prometheus
- Logz.io
- Kafka

## Configuration:

Create a `pump.conf` file:

```.json
{
  "analytics_storage_type": "redis",
  "analytics_storage_config": {
    "type": "redis",
    "host": "localhost",
    "port": 6379,
    "hosts": null,
    "username": "",
    "password": "",
    "database": 0,
    "optimisation_max_idle": 100,
    "optimisation_max_active": 0,
    "enable_cluster": false,
    "redis_use_ssl": false,
    "redis_ssl_insecure_skip_verify": false
  },
  "purge_delay": 1,
  "health_check_endpoint_name": "hello",
  "health_check_endpoint_port": 8083,
  "pumps": {
    "dummy": {
      "type": "dummy",
      "meta": {
        
      }
    },
    "mongo": {
      "type": "mongo",
      "meta": {
        "collection_name": "tyk_analytics",
        "mongo_url": "mongodb://username:password@{hostname:port},{hostname:port}/{db_name}"
      }
    },
    "mongo-pump-aggregate": {
      "type": "mongo-pump-aggregate",
      "meta": {
	"mongo_url": "mongodb://username:password@{hostname:port},{hostname:port}/{db_name}",
	"use_mixed_collection": true,
	"store_analytics_per_minute": false
      }
    },
    "csv": {
      "type": "csv",
      "meta": {
        "csv_dir": "./"
      }
    },
    "elasticsearch": {
      "type": "elasticsearch",
      "meta": {
        "index_name": "tyk_analytics",
        "elasticsearch_url": "http://localhost:9200",
        "enable_sniffing": false,
        "document_type": "tyk_analytics",
        "rolling_index": false,
        "extended_stats": false,
        "version": "5",
        "bulk_config":{
          "workers": 2,
          "flush_interval": 60
        }
      }
    },
    "influx": {
      "type": "influx",
      "meta": {
        "database_name": "tyk_analytics",
        "address": "http//localhost:8086",
        "username": "root",
        "password": "root",
        "fields": [
          "request_time"
        ],
        "tags": [
          "path",
          "response_code",
          "api_key",
          "api_version",
          "api_name",
          "api_id",
          "raw_request",
          "ip_address",
          "org_id",
          "oauth_id"
        ]
      }
    },
    "moesif": {
      "type": "moesif",
      "meta": {
        "application_id": ""
      }
    },
    "splunk": {
      "type": "splunk",
      "meta": {
        "collector_token": "<token>",
        "collector_url": "<url>",
        "ssl_insecure_skip_verify": false,
        "ssl_cert_file": "<cert-path>",
        "ssl_key_file": "<key-path>",
        "ssl_server_name": "<server-name>"
      }
    },
    "statsd": {
      "type": "statsd",
      "meta": {
        "address": "localhost:8125",
        "fields": [
          "request_time"
        ],
        "tags": [
          "path",
          "response_code",
          "api_key",
          "api_version",
          "api_name",
          "api_id",
          "raw_request",
          "ip_address",
          "org_id",
          "oauth_id"
        ]
      }
    },
    "dogstatsd": {
      "type": "dogstatsd",
      "meta": {
        "address": "localhost:8125",
        "namespace": "pump",
        "async_uds": true,
        "async_uds_write_timeout_seconds": 2,
        "buffered": true,
        "buffered_max_messages": 32
      }
    },
    "prometheus": {
      "type": "prometheus",
      "meta": {
        "listen_address": "localhost:9090",
        "path": "/metrics"
      }
    },
    "graylog": {
      "type": "graylog",
      "meta": {
        "host": "10.60.6.15",
        "port": 12216,
        "tags": [
          "method",
          "path",
          "response_code",
          "api_key",
          "api_version",
          "api_name",
          "api_id",
          "org_id",
          "oauth_id",
          "raw_request",
          "request_time",
          "raw_response"
        ]
      }
    },
    "hybrid": {
      "type": "hybrid",
      "meta": {
        "rpc_key": "5b5fd341e6355b5eb194765e",
        "api_key": "008d6d1525104ae77240f687bb866974",
        "connection_string": "localhost:9090",
	"aggregated": false,
        "use_ssl": false,
        "ssl_insecure_skip_verify": false,
        "group_id": "",
        "call_timeout": 30,
        "ping_timeout": 60,
        "rpc_pool_size": 30
      }
    },
    "logzio": {
      "type": "logzio",
      "meta": {
        "token": "<YOUR-LOGZ.IO-TOKEN>"
      }
    },
    "kafka": {
      "type": "kafka",
      "meta": {
        "broker": [
            "localhost:9092"
        ],
	"topic": "tyk-pump",
        "use_ssl": true,
        "ssl_insecure_skip_verify": false,
        "client_id": "tyk-pump",
        "timeout": 60,
        "compressed": true,
        "meta_data": {
            "key": "value"
        }
      }
    },
    "syslog": {
      "name": "syslog",
      "meta": {
        "transport": "udp",
        "network_addr": "localhost:5140",
        "log_level": 6,
        "tag":"syslog-pump"
      }
    }
  },
  "uptime_pump_config": {
    "collection_name": "tyk_uptime_analytics",
    "mongo_url": "mongodb://username:password@{hostname:port},{hostname:port}/{db_name}"
  },
  "dont_purge_uptime_data": false,
  "omit_detailed_recording": false,
  "obfuscate_keys":false
}
```

Settings are the same as for the original `tyk.conf` for redis and for mongoDB.

### Filter Records

This feature adds a new configuration field in each pump called filters and its structure is the following:
```json
"filters":{
  "api_ids":[],
  "org_ids":[],
  "response_codes":[],
  "skip_api_ids":[],
  "skip_org_ids":[],
  "skip_response_codes":[]
}
```
The fields api_ids, org_ids and response_codes works as allow list (APIs and orgs where we want to send the analytics records) and the fields skip_api_ids, skip_org_ids and skip_response_codes works as block list.

The priority is always block list configurations over allow list.

An example of configuration would be:
```json
"csv": {
 "type": "csv",
 "filters": {
   "org_ids": ["org1","org2"]
 },
 "meta": {
   "csv_dir": "./bar"
 }
}
```


### Environment Variables

Environment variables can be used to override the settings defined in the configuration file. See [Environment Variables](https://tyk.io/docs/tyk-configuration-reference/environment-variables/) in our docs for details. Where an environment variable is specified, its value will take precedence over the value in the configuration file.

### analytics_storage_config
```json
  "analytics_storage_config": {
    "type": "redis",
    "host": "localhost",
    "port": 6379,
    "hosts": null,
    "username": "",
    "password": "",
    "database": 0,
    "optimisation_max_idle": 100,
    "optimisation_max_active": 0,
    "enable_cluster": false,
    "redis_use_ssl": false,
    "redis_ssl_insecure_skip_verify": false
  },
```
`redis_use_ssl` - Setting this to true to use SSL when connecting to Redis

`redis_ssl_insecure_skip_verify` - Set this to true to tell Pump to ignore Redis' cert validation

### Uptime Data

`dont_purge_uptime_data` - Setting this to false will create a pump that pushes uptime data to MongoDB, so the Dashboard can read it. Disable by setting to true

### Omit Detailed Recording

`omit_detailed_recording` - Setting this to true will avoid writing raw_request and raw_response fields for each request in pumps. Defaults to false.

### Obfuscate Keys

`obfuscate_keys` - Setting this to true will obfuscate the API KEY from each record. 

### Health Check

From v2.9.4, we have introduced a `/health` endpoint to confirm the Pump is running. You need to configure the following settings:

- `health_check_endpoint_name` - The default is "hello" 
- `health_check_endpoint_port` - The default port is 8083

This returns a HTTP 200 OK response if the Pump is running.

### Tyk Dashboard

The Tyk Dashboard uses the "mongo-pump-aggregate" collection to display analytics.  This is different than the standard "mongo" pump plugin that will store individual analytic items into mongo.  The aggregate functionality was built to be fast, as querying raw analytics is expensive in large data sets.

### Elasticsearch Config

`"index_name"` - The name of the index that all the analytics data will be placed in. Defaults to "tyk_analytics"

`"elasticsearch_url"` - If sniffing is disabled, the URL that all data will be sent to. Defaults to "http://localhost:9200"

`"enable_sniffing"` - If sniffing is enabled, the "elasticsearch_url" will be used to make a request to get a list of all the nodes in the cluster, the returned addresses will then be used. Defaults to false

`"document_type"` - The type of the document that is created in ES. Defaults to "tyk_analytics"

`"rolling_index"` - Appends the date to the end of the index name, so each days data is split into a different index name. E.g. tyk_analytics-2016.02.28 Defaults to false

`"extended_stats"` - If set to true will include the following additional fields: Raw Request, Raw Response and User Agent.

`"version"` - Specifies the ES version. Use "3" for ES 3.X, "5" for ES 5.X, "6" for ES 6.X. Defaults to "3".

`"disable_bulk"` - Disable batch writing. Defaults to false.

`bulk_config`: Batch writing trigger configuration. Each option is an OR with eachother:
  * `wokers`: Number of workers. Defaults to 1.
  * `flush_interval`: Specifies the time in seconds to flush the data and send it to ES. Default disabled.
  * `bulk_actions`: Specifies the number of requests needed to flush the data and send it to ES. Defaults to 1000 requests. If it is needed, can be disabled with -1.
  * `bulk_size`: Specifies the size (in bytes) needed to flush the data and send it to ES. Defaults to 5MB. If it is needed, can be disabled with -1.

### Moesif Config
[Moesif](https://www.moesif.com/?language=tyk-api-gateway) is a user-centric API analytics and monitoring service for APIs. [More Info on Moesif for Tyk](https://www.moesif.com/solutions/track-api-program?language=tyk-api-gateway)

- `"application_id"` - Moesif App Id JWT. Multiple api_id's will go under the same app id.
- `"request_header_masks"` - (optional) An option to mask a specific request header field. Type: String Array `[] string`
- `"request_body_masks"` - (optional) An option to mask a specific - request body field. Type: String Array `[] string`
- `"response_header_masks"` - (optional) An option to mask a specific response header field. Type: String Array `[] string`
- `"response_body_masks"` - (optional) An option to mask a specific response body field. Type: String Array `[] string`
- `"disable_capture_request_body"` - (optional) An option to disable logging of request body. Type: Boolean. Default value is `false`.
- `"disable_capture_response_body"` - (optional) An option to disable logging of response body. Type: Boolean. Default value is `false`.
- `"user_id_header"` - (optional) An optional field name to identify User from a request or response header. Type: String.
- `"company_id_header"` - (optional) An optional field name to identify Company (Account) from a request or response header. Type: String.

### Hybrid RPC Config

Hybrid Pump allows you to install Tyk Pump inside Multi-Cloud or MDCB Worker installations. You can configure Tyk Pump to send data to the source of your choice (i.e. ElasticSearch), and in parallel, forward analytics to the Tyk Cloud. Additionally, you can set the aggregated flag to send only aggregated analytics to MDCB or Tyk Cloud, in order to save network bandwidth between DCs.

NOTE: Make sure your tyk.conf has analytics_config.type set to empty string value.

rpc_key - Put your organization ID in this field.

api_key - This the API key of a user used to authenticate and authorise the Gateway’s access through MDCB. The user should be a standard Dashboard user with minimal privileges so as to reduce risk if compromised. The suggested security settings are read for Real-time notifications and the remaining options set to deny.

aggregated - Set this field to true to send only aggregated analytics to MDCB or Tyk Cloud.

connection_string - The MDCB instance or load balancer.

use_ssl - Set this field to true if you need secured connection (default value is false).

ssl_insecure_skip_verify - Set this field to true if you use self signed certificate.

group_id - This is the “zone” that this instance inhabits, e.g. the DC it lives in. It must be unique to each slave cluster / DC.

call_timeout - This is the timeout (in milliseconds) for RPC calls.

rpc_pool_size - This is maximum number of connections to MDCB.

### Prometheus
Prometheus is an open-source monitoring system with a dimensional data model, flexible query language, efficient time series database and modern alerting approach.

Add the following section to expose "/metrics" endpoint:
```.json
"prometheus": {
  "type": "prometheus",
	"meta": {
		"listen_address": "localhost:9090",
		"path": "/metrics"
	}
},
```

`Note` - When run as docker image then `"listen_address": ":9090"`

Tyk expose the following counters:
- tyk_http_status{code, api}
- tyk_http_status_per_path{code, api, path, method}
- tyk_http_status_per_key{code, key}
- tyk_http_status_per_oauth_client{code, client_id}

And the following Histogram for latencies:
- tyk_latency{type, api}

### DogStatsD

- `address`: address of the datadog agent including host & port
- `namespace`: prefix for your metrics to datadog
- `async_uds`: Enable async UDS over UDP https://github.com/Datadog/datadog-go#unix-domain-sockets-client
- `async_uds_write_timeout_seconds`: Integer write timeout in seconds if `async_uds: true`
- `buffered`: Enable buffering of messages
- `buffered_max_messages`: Max messages in single datagram if `buffered: true`. Default 16
- `sample_rate`: default 1 which equates to 100% of requests. To sample at 50%, set to 0.5

```.json
"dogstatsd": {
  "type": "dogstatsd",
  "meta": {
    "address": "localhost:8125",
    "namespace": "pump",
    "async_uds": true,
    "async_uds_write_timeout_seconds": 2,
    "buffered": true,
    "buffered_max_messages": 32,
    "sample_rate": 0.5
  }
},
```

On startup, you should see the loaded configs when initializing the dogstatsd pump
```
[May 10 15:23:44]  INFO dogstatsd: initializing pump
[May 10 15:23:44]  INFO dogstatsd: namespace: pump.
[May 10 15:23:44]  INFO dogstatsd: sample_rate: 50%
[May 10 15:23:44]  INFO dogstatsd: buffered: true, max_messages: 32
[May 10 15:23:44]  INFO dogstatsd: async_uds: true, write_timeout: 2s
```
### Splunk Config

Setting up Splunk with a *HTTP Event Collector*

- `collector_token`: address of the datadog agent including host & port
- `collector_url`: endpoint the Pump will send analytics too.  Should look something like:

`https://splunk:8088/services/collector/event`

- `ssl_insecure_skip_verify`: Controls whether the pump client verifies the Splunk server's certificate chain and host name.

Example:
```json
    "splunk": {
      "type": "splunk",
      "meta": {
        "collector_token": "<token>",
        "collector_url": "<url>",
        "ssl_insecure_skip_verify": false,
        "ssl_cert_file": "<cert-path>",
        "ssl_key_file": "<key-path>",
        "ssl_server_name": "<server-name>"
      }
    },
```

### Logzio Config

Logz.io is a cloud observability platform providing Log Management built on ELK, Infrastructure Monitoring based on Grafana, and an ELK-based Cloud SIEM.

The following configuration values are available:

Example simplest configuration just needs the token for sending data to your logzio account.

```.json
"logzio": {
      "type": "logzio",
      "meta": {
        "token": "<YOUR-LOGZ.IO-TOKEN>"
      }
    }
```

More advanced fields:

`meta.url` - If you do not want to use the default Logzio url i.e. when using a proxy. Default is `https://listener.logz.io:8071`
`meta.queue_dir` - The directory for the queue.
`meta.drain_duration` - Set drain duration (flush logs on disk). Default value is `3s`
`meta.disk_threshold` - Set disk queue threshold, once the threshold is crossed the sender will not enqueue the received logs. Default value is `98` (percentage of disk).
`meta.check_disk_space` - Set the sender to check if it crosses the maximum allowed disk usage. Default value is `true`.


### Kafka Config

* `broker`: The list of brokers used to discover the partitions available on the kafka cluster. E.g. "localhost:9092"
* `use_ssl`: Enables SSL connection. 
* `ssl_insecure_skip_verify`: Controls whether the pump client verifies the kafka server's certificate chain and host name.
* `client_id`: Unique identifier for client connections established with Kafka.
* `topic`: The topic that the writer will produce messages to.
* `timeout`: Timeout is the maximum amount of time will wait for a connect or write to complete. 
* `compressed`: Enable "github.com/golang/snappy" codec to be used to compress Kafka messages. By default is false
* `meta_data`: Can be used to set custom metadata inside the kafka message
* `ssl_cert_file`: Can be used to set custom certificate file for authentication with kafka.
* `ssl_key_file`: Can be used to set custom key file for authentication with kafka.


### Syslog
`"transport"` - Possible values are `udp, tcp, tls` in string form

`"network_addr"` - Host & Port combination of your syslog daemon ie: `"localhost:5140"`

`"log_level"` - The severity level, an integer from 0-7, based off the Standard: [Syslog Severity Levels](https://en.wikipedia.org/wiki/Syslog#Severity_level)

`"tag"` - Prefix tag

When working with FluentD, you should provide a [FluentD Parser](https://docs.fluentd.org/input/syslog) based on the OS you are using so that FluentD can correctly read the logs

```.json
"syslog": {
  "name": "syslog",
  "meta": {
    "transport": "udp",
    "network_addr": "localhost:5140",
    "log_level": 6,
    "tag": "syslog-pump"
  }
```

## Compiling & Testing

1. Download dependent packages:

```
go get -t -d -v ./...
```

2. Compile:

```
go build -v ./...
```

3. Test

```
go test -v ./...
```

### Multiple Pumps

From Tyk Pump v0.6.0 you can now create multiple pumps of the same type by by setting the top level type as a custom values. For example:

```{.json}
"csv": {
  "type": "csv",
  "meta": {
    "csv_dir": "./"
  }
},
"csv_alt": {
  "type": "csv",
    "meta": {
    "csv_dir": "./"
  }
}
```
