package prometheus

import (
	"fmt"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/Shopify/sarama"
	"github.com/VictoriaMetrics/VictoriaMetrics/lib/logger"
	"github.com/VictoriaMetrics/metrics"
	"github.com/gogo/protobuf/proto"
	"tools.adidas-group.com/bitbucket/TM/prom2kafka/app/prom2kafka/concurrencylimiter"
	"tools.adidas-group.com/bitbucket/TM/prom2kafka/lib/prompb"
)

var (
	samplesInserted      = metrics.NewCounter(`prom2kafka_samples_received_total`)
	kafkaSamplesInserted = metrics.NewCounter(`prom2kafka_kafka_samples_forwarded_total`)
	samplesPerRequest    = metrics.NewSummary(`prom2kafka_samples_per_request`)
	latency              = metrics.NewHistogram(`prom2kafka_request_duration_seconds`)
	kafkaLatency         = metrics.NewHistogram(`prom2kafka_kafka_write_duration_seconds`)
)

// InsertHandler processes remote write for prometheus.
func InsertHandler(r *http.Request, maxSize int64, timeout time.Duration, topic string, newKafkaClient func() (sarama.SyncProducer, error)) error {
	defer latency.UpdateDuration(time.Now())
	return concurrencylimiter.Do(func() error {
		return insertHandlerInternal(r, maxSize, timeout, topic, newKafkaClient)
	})
}

func insertHandlerInternal(r *http.Request, maxSize int64, timeout time.Duration, topic string, newKafkaClient func() (sarama.SyncProducer, error)) error {
	ctx := getPushCtx(newKafkaClient)
	defer putPushCtx(ctx)
	if err := ctx.Read(r, maxSize); err != nil {
		return err
	}
	timeseries := ctx.req.Timeseries

	rowsTotal := 0
	ik := make([]*sarama.ProducerMessage, 0, len(timeseries))
	for i := range timeseries {
		ts := &timeseries[i]

		// hack: *prompb.Labels implements Marshaling but *[]prompb.Label does not
		key, err := proto.Marshal((*prompb.Labels)(&ts.Labels))
		if err != nil {
			logger.Errorf("marshal error for kafka labels: %s", err)
			continue
		}
		for _, s := range ts.Samples {
			value, err := proto.Marshal(&s)
			if err != nil {
				logger.Errorf("marshal error for kafka value: %s", err)
				continue
			}
			ik = append(ik, &sarama.ProducerMessage{
				Topic: topic,
				Key:   sarama.ByteEncoder(key),
				Value: sarama.ByteEncoder(value),
			})

		}

		rowsTotal += len(ts.Samples)
	}
	samplesInserted.Add(rowsTotal)
	samplesPerRequest.Update(float64(rowsTotal))

	if len(ik) > 0 {
		now := time.Now()
		err := ctx.kafka.SendMessages(ik)
		kafkaLatency.UpdateDuration(now)
		if err != nil {
			logger.Errorf("error sending metrics to kafka: %s", err)
		}
		kafkaSamplesInserted.Add(len(ik))
	}

	return nil
}

type pushCtx struct {
	req    prompb.WriteRequest
	reqBuf []byte
	kafka  sarama.SyncProducer
}

func (ctx *pushCtx) reset() {
	ctx.req.Reset()
	ctx.reqBuf = ctx.reqBuf[:0]
}

func (ctx *pushCtx) Read(r *http.Request, maxSize int64) error {
	prometheusReadCalls.Inc()

	var err error
	ctx.reqBuf, err = prompb.ReadSnappy(ctx.reqBuf[:0], r.Body, maxSize)
	if err != nil {
		prometheusReadErrors.Inc()
		return fmt.Errorf("cannot read prompb.WriteRequest: %s", err)
	}
	if err = ctx.req.Unmarshal(ctx.reqBuf); err != nil {
		prometheusUnmarshalErrors.Inc()
		return fmt.Errorf("cannot unmarshal prompb.WriteRequest with size %d bytes: %s", len(ctx.reqBuf), err)
	}
	return nil
}

var (
	prometheusReadCalls       = metrics.NewCounter(`prom2kafka_read_calls_total`)
	prometheusReadErrors      = metrics.NewCounter(`prom2kafka_read_errors_total`)
	prometheusUnmarshalErrors = metrics.NewCounter(`prom2kafka_unmarshal_errors_total`)
)

func getPushCtx(newKafkaClient func() (sarama.SyncProducer, error)) *pushCtx {
	select {
	case ctx := <-pushCtxPoolCh:
		return ctx
	default:
		if v := pushCtxPool.Get(); v != nil {
			return v.(*pushCtx)
		}
		buf := &pushCtx{}
		buf.reqBuf = make([]byte, 0, 1024)
		var err error
		buf.kafka, err = newKafkaClient()
		if err != nil {
			logger.Panicf("unable to create KafkaClient: %s", err)
		}
		return buf
	}
}

func putPushCtx(ctx *pushCtx) {
	ctx.reset()
	select {
	case pushCtxPoolCh <- ctx:
	default:
		pushCtxPool.Put(ctx)
	}
}

var pushCtxPool sync.Pool
var pushCtxPoolCh = make(chan *pushCtx, runtime.GOMAXPROCS(-1))
