go-metrics-reporter
====================

This is a reporter for the [go-metrics](https://github.com/rcrowley/go-metrics) library which will post the metrics to [InfluxDB](https://influxdb.com/).

This version adds a measurement for the metrics, moves the histogram bucket names into tags, similar to the behavior of hitograms in telegraf, and aligns all metrics in a batch on the same timestamp.

Additionally, metrics can be aligned to the beginning of a bucket as defined by the interval.

Setting align to true will cause the timestamp to be truncated down to the nearest even integral of the reporting interval.

For example, if the interval is 30 seconds, tiemstamps will be aligned on :00 and :30 for every reporting interval.

This also maps to a similar option in Telegraf.

Note
----

This is only compatible with InfluxDB 2.X ([see details](https://github.com/influxdata/influxdb-client-go)).

Please use [go-metrics-influxdb](https://github.com/vrischmann/go-metrics-influxdb) with InfluxDB 1.X.

Usage
-----

```go
import "github.com/teamlint/go-metrics-reporter"

r := reporter.New(
    metrics.DefaultRegistry,    // go-metrics registry
    serverURL,                  // the InfluxDB url
    token,                      // your InfluxDB authentication token
    org,                        // your InfluxDB org
    bucket,                     // writes to desired bucket
    interval,                   // interval
)
defer r.Close()
go r.Run()
````

License
-------

go-metrics-reporter is licensed under the MIT license. See the LICENSE file for details.
