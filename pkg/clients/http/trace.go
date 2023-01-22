package http

import (
	"context"
	"net/http/httptrace"
	"sync"
	"time"

	"github.com/antonpriyma/otus-highload/pkg/log"
)

type traceDumper interface {
	DumpTrace(ctx context.Context, logger log.Logger)
}

type dummyTracer struct{}

func (t dummyTracer) DumpTrace(ctx context.Context, logger log.Logger) {}

type traceRecord struct {
	Fields  log.Fields
	Message string
}

type traceStorage struct {
	Mu sync.Mutex

	Start   time.Time
	Records []traceRecord
}

func (s *traceStorage) Add(rec traceRecord) {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	s.Records = append(s.Records, rec)
}

func (s *traceStorage) DumpTrace(ctx context.Context, logger log.Logger) {
	logger = logger.ForCtx(ctx).WithFields(log.Fields{
		"from_httptrace":     true,
		"context_created_at": s.Start.Format(shortNano),
	})

	for _, rec := range s.Records {
		logger.WithFields(rec.Fields).Info(rec.Message)
	}
}

// minutes, seconds, nanoseconds
const shortNano = "04:05.999999999"

func contextWithConnectTrace(ctx context.Context) (context.Context, traceDumper) {
	trace := traceStorage{
		Start:   time.Now(),
		Records: make([]traceRecord, 0, 6),
	}

	tm := trace.Start
	return httptrace.WithClientTrace(
		ctx,
		&httptrace.ClientTrace{
			GetConn: func(hostPort string) {
				newTm := time.Now()

				trace.Add(traceRecord{
					Fields: log.Fields{
						"host_port":   hostPort,
						"action_time": time.Now().Format(shortNano),
						"duration":    newTm.Sub(tm),
					},
					Message: "get conn from idle pool",
				})

				tm = newTm
			},

			DNSStart: func(info httptrace.DNSStartInfo) {
				newTm := time.Now()

				trace.Add(traceRecord{
					Fields: log.Fields{
						"host":        info.Host,
						"action_time": time.Now().Format(shortNano),
						"duration":    newTm.Sub(tm),
					},
					Message: "dns start",
				})

				tm = newTm
			},

			DNSDone: func(info httptrace.DNSDoneInfo) {
				newTm := time.Now()

				trace.Add(traceRecord{
					Fields: log.Fields{
						"addrs":       info.Addrs,
						"err":         info.Err,
						"coalesced":   info.Coalesced,
						"action_time": time.Now().Format(shortNano),
						"duration":    newTm.Sub(tm),
					},
					Message: "dns done",
				})

				tm = newTm
			},

			ConnectStart: func(network string, addr string) {
				newTm := time.Now()

				trace.Add(traceRecord{
					Fields: log.Fields{
						"network":     network,
						"addr":        addr,
						"action_time": time.Now().Format(shortNano),
						"duration":    newTm.Sub(tm),
					},
					Message: "connect start",
				})

				tm = newTm
			},

			ConnectDone: func(network string, addr string, err error) {
				newTm := time.Now()

				trace.Add(traceRecord{
					Fields: log.Fields{
						"network":     network,
						"addr":        addr,
						"err":         err,
						"action_time": time.Now().Format(shortNano),
						"duration":    newTm.Sub(tm),
					},
					Message: "connect done",
				})

				tm = newTm
			},

			GotConn: func(cn httptrace.GotConnInfo) {
				newTm := time.Now()

				trace.Add(traceRecord{
					Fields: log.Fields{
						"reused":      cn.Reused,
						"was_idle":    cn.WasIdle,
						"idle_time":   cn.IdleTime,
						"addr":        cn.Conn.RemoteAddr(),
						"action_time": time.Now().Format(shortNano),
						"duration":    newTm.Sub(tm),
					},
					Message: "got connect",
				})

				tm = newTm
			},
		},
	), &trace
}
