package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/Shopify/sarama"
	"github.com/VictoriaMetrics/VictoriaMetrics/lib/buildinfo"
	"github.com/VictoriaMetrics/VictoriaMetrics/lib/httpserver"
	"github.com/VictoriaMetrics/VictoriaMetrics/lib/logger"
	"github.com/VictoriaMetrics/VictoriaMetrics/lib/procutil"
	"github.com/VictoriaMetrics/metrics"
	"tools.adidas-group.com/bitbucket/TM/prom2kafka/app/prom2kafka/concurrencylimiter"
	"tools.adidas-group.com/bitbucket/TM/prom2kafka/app/prom2kafka/prometheus"
)

var (
	caFile               = flag.String("ca-file", "/etc/ca.crt", "CA certificate file to validate kafka certs.")
	certFile             = flag.String("cert-file", "/etc/cert.pem", "Certificate file to identify against kafka.")
	keyFile              = flag.String("key-file", "/etc/key.pem", "Key file to identify against kafka.")
	kafkaBrokers         = flag.String("kafka-brokers", "", "CSV list of broker host or host:port of the kafka cluster to send samples to.")
	kafkaTopic           = flag.String("kafka-topic", "", "Kafka topic to write samples to.")
	httpListenAddr       = flag.String("httpListenAddr", ":9201", "Address to listen for http connections.")
	compression          = flag.String("compression", "zstd", "Compression codec to use for compressing messages to Kafka. Valid values: none, snappy, lz4, gzip, zstd.")
	timeout              = flag.Duration("timeout", 5*time.Second, "Timeout to use when sending samples.")
	batchSize            = flag.Int("batch-size", 1000, "The batch size to produce into kafka.")
	maxInsertRequestSize = flag.Int("maxInsertRequestSize", 32*1024*1024, "The maximum size of a single insert request in bytes.")
)

func main() {
	flag.Parse()
	buildinfo.Init()
	logger.Init()
	var err error
	logger.Infof("initializing TLS configuration...")
	tlsConfig := &tls.Config{}
	if *caFile != "" {
		caCert, err := ioutil.ReadFile(*caFile)
		if err != nil {
			logger.Fatalf("couldn't load CA certificate %s: %s", *caFile, err)
			return
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
		tlsConfig.RootCAs = caCertPool
	}

	if *certFile != "" && *keyFile != "" {
		cert, err := tls.LoadX509KeyPair(*certFile, *keyFile)
		if err != nil {
			logger.Fatalf("couldn't load tls certificate %s, %s: %s", *certFile, *keyFile, err)
			return
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}
	logger.Infof("successfully initialized TLS configuration")

	logger.Infof("initializing kafka producer...")

	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll // Wait for all in-sync replicas to ack the message
	config.Producer.Retry.Max = 10                   // Retry up to 10 times to produce the message
	config.Producer.Return.Successes = true
	if tlsConfig != nil {
		config.Net.TLS.Config = tlsConfig
		config.Net.TLS.Enable = true
	}
	version, err := sarama.ParseKafkaVersion("2.3.0")
	if err != nil {
		logger.Fatalf("Error parsing Kafka version: %v", err)
	}
	config.Version = version
	// On the broker side, you may want to change the following settings to get
	// stronger consistency guarantees:
	// - For your broker, set `unclean.leader.election.enable` to false
	// - For the topic, you could increase `min.insync.replicas`.

	switch *compression {
	case "snappy":
		config.Producer.Compression = sarama.CompressionSnappy
	case "gzip":
		config.Producer.Compression = sarama.CompressionGZIP
	case "lz4":
		config.Producer.Compression = sarama.CompressionLZ4
	case "zstd":
		config.Producer.Compression = sarama.CompressionZSTD
	case "none":
		config.Producer.Compression = sarama.CompressionNone
	default:
		logger.Fatalf("must specify a valid compression codec")
	}
	newKafkaClient = func() (sarama.SyncProducer, error) {
		return sarama.NewSyncProducer(strings.Split(*kafkaBrokers, ","), config)
	}

	concurrencylimiter.Init()

	go func() {
		httpserver.Serve(*httpListenAddr, requestHandler)
	}()

	sig := procutil.WaitForSigterm()
	logger.Infof("service received signal %s", sig)

	logger.Infof("gracefully shutting down the service at %q", *httpListenAddr)
	startTime := time.Now()
	if err := httpserver.Stop(*httpListenAddr); err != nil {
		logger.Fatalf("cannot stop the service: %s", err)
	}
	logger.Infof("successfully shut down the service in %s", time.Since(startTime))

	logger.Infof("the vminsert has been stopped")
}

func requestHandler(w http.ResponseWriter, r *http.Request) bool {
	switch r.URL.Path {
	case "/ready":
		if concurrencylimiter.Overloaded() {
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte("NOT-OK"))
			return true
		}
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
		return true
	default:
		prometheusWriteRequests.Inc()
		if err := prometheus.InsertHandler(r, int64(*maxInsertRequestSize), *timeout, *kafkaTopic, newKafkaClient); err != nil {
			prometheusWriteErrors.Inc()
			httpserver.Errorf(w, "error in %q: %s", r.URL.Path, err)
			return true
		}
		w.WriteHeader(http.StatusNoContent)
		return true
	}
}

var (
	prometheusWriteRequests = metrics.NewCounter(`prom2kafka_http_requests_total`)
	prometheusWriteErrors   = metrics.NewCounter(`prom2kafka_http_request_errors_total`)
)

var newKafkaClient func() (sarama.SyncProducer, error)
