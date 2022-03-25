package reporter

import (
	"log"
	"net/url"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
	"github.com/rcrowley/go-metrics"
)

type reporter struct {
	reg      metrics.Registry
	align    bool
	dburl    url.URL
	org      string
	bucket   string
	interval time.Duration

	measurement string
	tags        map[string]string

	client influxdb2.Client
}

func New(r metrics.Registry, serverURL string, token string, org string, bucket string, interval time.Duration) *reporter {
	return NewWithTags(r, serverURL, token, org, bucket, "reporter", interval, map[string]string{}, true)
}

func NewWithTags(r metrics.Registry, serverURL string, token string, org string, bucket string, measurement string, interval time.Duration, tags map[string]string, align bool) *reporter {
	// Create a client
	client := influxdb2.NewClient(serverURL, token)

	return &reporter{
		reg:         r,
		interval:    interval,
		align:       align,
		bucket:      bucket,
		measurement: measurement,
		org:         org,
		tags:        tags,
		client:      client,
	}
}

func (r *reporter) Close() {
	r.client.Close()
}

func (r *reporter) write(pts []*write.Point) {
	// get non-blocking write client
	writeAPI := r.client.WriteAPI(r.org, r.bucket)

	for _, p := range pts {
		writeAPI.WritePoint(p)
	}
	// Flush writes
	writeAPI.Flush()
}

func (r *reporter) Run() {
	intervalTicker := time.Tick(r.interval)
	for {
		select {
		case <-intervalTicker:
			if err := r.send(); err != nil {
				log.Printf("unable to send metrics to InfluxDB. err=%v", err)
			}
		}
	}
}

func (r *reporter) send() error {
	var pts []*write.Point

	now := time.Now()
	if r.align {
		now = now.Truncate(r.interval)
	}
	r.reg.Each(func(name string, i interface{}) {

		switch metric := i.(type) {
		case metrics.Counter:
			ms := metric.Snapshot()
			pts = append(pts,
				influxdb2.NewPoint(
					r.measurement,
					r.tags,
					map[string]interface{}{
						name: ms.Count(),
					},
					now,
				))
		case metrics.Gauge:
			ms := metric.Snapshot()
			pts = append(pts,
				influxdb2.NewPoint(
					r.measurement,
					r.tags,
					map[string]interface{}{
						name: ms.Value(),
					},
					now,
				))
		case metrics.GaugeFloat64:
			ms := metric.Snapshot()
			pts = append(pts, influxdb2.NewPoint(
				r.measurement,
				r.tags,
				map[string]interface{}{
					name: ms.Value(),
				},
				now,
			))
		case metrics.Histogram:
			ms := metric.Snapshot()
			ps := ms.Percentiles([]float64{0.5, 0.75, 0.95, 0.99})
			fields := map[string]float64{
				"count":    float64(ms.Count()),
				"max":      float64(ms.Max()),
				"mean":     ms.Mean(),
				"min":      float64(ms.Min()),
				"stddev":   ms.StdDev(),
				"variance": ms.Variance(),
				"p50":      ps[0],
				"p75":      ps[1],
				"p95":      ps[2],
				"p99":      ps[3],
			}
			for k, v := range fields {
				pts = append(pts, influxdb2.NewPoint(
					r.measurement,
					bucketTags(k, r.tags),
					map[string]interface{}{
						name: v,
					},
					now,
				))

			}
		case metrics.Meter:
			ms := metric.Snapshot()
			fields := map[string]float64{
				"count": float64(ms.Count()),
				"m1":    ms.Rate1(),
				"m5":    ms.Rate5(),
				"m15":   ms.Rate15(),
				"mean":  ms.RateMean(),
			}
			for k, v := range fields {
				pts = append(pts,
					influxdb2.NewPoint(
						r.measurement,
						bucketTags(k, r.tags),
						map[string]interface{}{
							name: v,
						},
						now,
					))
			}

		case metrics.Timer:
			ms := metric.Snapshot()
			ps := ms.Percentiles([]float64{0.5, 0.75, 0.95, 0.99})
			fields := map[string]float64{
				"count":    float64(ms.Count()),
				"max":      float64(ms.Max()),
				"mean":     ms.Mean(),
				"min":      float64(ms.Min()),
				"stddev":   ms.StdDev(),
				"variance": ms.Variance(),
				"p50":      ps[0],
				"p75":      ps[1],
				"p95":      ps[2],
				"p99":      ps[3],
				"m1":       ms.Rate1(),
				"m5":       ms.Rate5(),
				"m15":      ms.Rate15(),
				"meanrate": ms.RateMean(),
			}
			for k, v := range fields {
				pts = append(pts, influxdb2.NewPoint(
					r.measurement,
					bucketTags(k, r.tags),
					map[string]interface{}{
						name: v,
					},
					now,
				))
			}
		}
	})

	r.write(pts)
	return nil
}
func bucketTags(bucket string, tags map[string]string) map[string]string {
	m := map[string]string{}
	for tk, tv := range tags {
		m[tk] = tv
	}
	m["bucket"] = bucket
	return m
}
