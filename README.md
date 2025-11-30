# go-musthave-metrics-tpl

Server and agent for collecting metrics and alerting.

# Benchmarks and optimisation

Бенчмарк проводился с помощью утилиты `hey`.

## Бенчмарк

Команда нагружает сервер

- 2000 POST-запросами,
- 50 потоков одновременно,
- каждый отправляет gzip-запакованное JSON-тело,
- плюс обязательный SHA256-хеш для проверки целостности.

```bash
hey -m POST -n 2000 -c 50 -disable-compression \
  -D profiles/post_updates_body.gz \
  -H "Accept: application/json" \
  -H "Content-Encoding: gzip" \
  -H "Content-Type: application/json" \
  -H "Hashsha256: 00a74b9ead06b09fd3683158b77529b63507665ed218d9097fd8d768ab63ded8" \
  http://localhost:8081/updates/
```

## Allocations

```json
> go tool pprof -diff_base=base_allocs.pprof result.pprof

File: ___SERVER___DB___hashing
Type: alloc_space
Time: 2025-11-28 02:25:31 +05
Showing nodes accounting for -41MB, 14.34% of 285.87MB total
Dropped 8 nodes (cum <= 1.43MB)
flat  flat%   sum%        cum   cum%
-18.54MB  6.48%  6.48%   -18.54MB  6.48%  bytes.growSlice
15.55MB  5.44%  1.04%    15.55MB  5.44%  reflect.growslice
-9MB  3.15%  4.19%    -9.50MB  3.32%  github.com/rs/zerolog.(*Event).caller
-5.50MB  1.92%  6.12%   -30.54MB 10.68%  github.com/rs/zerolog.init.func7
-5.02MB  1.76%  7.88%    -5.54MB  1.94%  compress/flate.NewReader
-5MB  1.75%  9.62%   -23.54MB  8.23%  bytes.(*Buffer).grow
-5MB  1.75% 11.37%   -33.50MB 11.72%  github.com/MKhiriev/stunning-adventure/internal/store.(*DB).saveAllMetrics
4MB  1.40%  9.97%    -3.50MB  1.22%  github.com/jackc/pgx/v5/stdlib.(*Conn).ExecContext
-3.50MB  1.22% 11.20%       -6MB  2.10%  github.com/jackc/pgx/v5/pgconn.(*ResultReader).Read
-3MB  1.05% 12.25%    -7.50MB  2.62%  database/sql.resultFromStatement
-2.51MB  0.88% 13.13%    -2.51MB  0.88%  bufio.NewReaderSize (inline)
2.51MB  0.88% 12.25%     2.51MB  0.88%  io.ReadAll
-2.50MB  0.87% 13.12%    -2.50MB  0.87%  github.com/jackc/pgx/v5/pgconn.(*PgConn).makeCommandTag (inline)
2MB   0.7% 12.42%        2MB   0.7%  github.com/MKhiriev/stunning-adventure/internal/handlers.init.func2
-2MB   0.7% 13.12%       -2MB   0.7%  github.com/jackc/pgx/v5/pgconn/ctxwatch.(*ContextWatcher).Watch
1.50MB  0.53% 12.60%     1.50MB  0.53%  sync.(*Pool).pinSlow
-1.50MB  0.52% 13.12%   -48.53MB 16.98%  github.com/MKhiriev/stunning-adventure/internal/service.(*ValidatingMetricsService).SaveAll
1.02MB  0.36% 12.77%     1.02MB  0.36%  github.com/jackc/pgx/v5/pgtype.(*Map).buildReflectTypeToType (inline)
1.01MB  0.35% 12.41%     1.01MB  0.35%  encoding/json.(*Decoder).refill
-1MB  0.35% 12.76%       -1MB  0.35%  net/textproto.readMIMEHeader
1MB  0.35% 12.41%        1MB  0.35%  encoding/json.NewDecoder (inline)
-1MB  0.35% 12.76%       -1MB  0.35%  os.statNolog
-1MB  0.35% 13.11%       -1MB  0.35%  database/sql.driverArgsConnLocked
1MB  0.35% 12.76%        1MB  0.35%  github.com/jackc/pgx/v5/pgconn.(*PgConn).convertRowDescription
-1MB  0.35% 13.11%       -1MB  0.35%  context.(*cancelCtx).propagateCancel
-1MB  0.35% 13.46%       -1MB  0.35%  reflect.copyVal
-1MB  0.35% 13.81%       -1MB  0.35%  database/sql.(*Tx).grabConn
-1MB  0.35% 14.16%       -1MB  0.35%  reflect.New
-0.52MB  0.18% 14.34%    -0.52MB  0.18%  compress/flate.(*dictDecoder).init (inline)
-0.50MB  0.18% 14.52%    -0.50MB  0.18%  github.com/rs/zerolog/internal/json.Encoder.AppendString
0.50MB  0.18% 14.34%     0.50MB  0.18%  runtime.allocm
-0.50MB  0.17% 14.52%    -1.48MB  0.52%  github.com/jackc/pgx/v5/stdlib.(*driverConnector).Connect
-0.50MB  0.17% 14.69%    -0.50MB  0.17%  net/http.(*Request).WithContext (inline)
0.50MB  0.17% 14.52%     0.50MB  0.17%  github.com/go-chi/chi/v5.NewRouteContext
0.50MB  0.17% 14.34%     1.50MB  0.52%  database/sql.(*DB).prepareDC
-0.50MB  0.17% 14.52%       -2MB   0.7%  context.WithDeadlineCause
0.50MB  0.17% 14.34%     0.50MB  0.17%  crypto/internal/fips140/sha256.(*Digest).MarshalBinary
-0.50MB  0.17% 14.52%    -0.50MB  0.17%  time.newTimer
0.50MB  0.17% 14.34%     0.50MB  0.17%  context.withCancel (inline)
-0.50MB  0.17% 14.52%   -30.98MB 10.84%  github.com/MKhiriev/stunning-adventure/internal/handlers.(*Handler).WithHashing-fm.(*Handler).WithHashing.func1
-0.50MB  0.17% 14.69%    -0.50MB  0.17%  net.(*conn).Read
-0.50MB  0.17% 14.87%    -0.50MB  0.17%  crypto/internal/fips140/hmac.New[go.shape.interface { BlockSize int; Reset; Size int; Sum []uint8; Write  }]
0.50MB  0.17% 14.69%        1MB  0.35%  database/sql.(*driverConn).prepareLocked
-0.50MB  0.17% 14.87%    -0.50MB  0.17%  encoding/hex.EncodeToString
0.50MB  0.17% 14.69%     0.50MB  0.17%  bytes.NewReader
0.50MB  0.17% 14.52%     0.50MB  0.17%  encoding/json.(*decodeState).object
-0.50MB  0.17% 14.69%       -1MB  0.35%  github.com/jackc/pgx/v5/pgconn.(*PgConn).scramAuth
0.50MB  0.17% 14.52%     0.50MB  0.17%  time.map.init.0
0.50MB  0.17% 14.34%     0.50MB  0.17%  crypto/internal/fips140/sha256.(*Digest).Sum
-0.50MB  0.17% 14.52%    -0.50MB  0.17%  syscall.anyToSockaddr
-0.50MB  0.17% 14.69%    -0.50MB  0.17%  encoding/json.appendString[go.shape.string]
-0.50MB  0.17% 14.87%    -0.50MB  0.17%  net/textproto.(*Reader).ReadLine (inline)
0.50MB  0.17% 14.69%     0.50MB  0.17%  encoding/json.(*scanner).pushParseState
-0.50MB  0.17% 14.87%     0.50MB  0.17%  github.com/jackc/pgx/v5/stdlib.(*Conn).PrepareContext
0.50MB  0.17% 14.69%   -39.02MB 13.65%  net/http.(*conn).serve
0.50MB  0.17% 14.52%     0.50MB  0.17%  reflect.packEface
0.50MB  0.17% 14.34%     0.50MB  0.17%  net/http.(*conn).readRequest
0     0% 14.34%    -2.51MB  0.88%  bufio.NewReader (inline)
0     0% 14.34%   -23.04MB  8.06%  bytes.(*Buffer).Write
0     0% 14.34%    -0.50MB  0.18%  bytes.(*Buffer).WriteByte
0     0% 14.34%    -8.55MB  2.99%  compress/gzip.(*Reader).Reset
0     0% 14.34%    -5.54MB  1.94%  compress/gzip.(*Reader).readHeader
0     0% 14.34%   -94.69MB 33.12%  compress/gzip.NewReader (inline)
0     0% 14.34%     0.50MB  0.17%  context.WithCancel
0     0% 14.34%       -2MB   0.7%  context.WithDeadline (inline)
0     0% 14.34%       -2MB   0.7%  context.WithTimeout
0     0% 14.34%    -0.50MB  0.17%  crypto/hmac.New
0     0% 14.34%     0.50MB  0.17%  crypto/internal/fips140/hmac.(*HMAC).Reset
0     0% 14.34%     0.50MB  0.17%  crypto/internal/fips140/hmac.(*HMAC).Sum
0     0% 14.34%       -2MB   0.7%  database/sql.(*DB).BeginTx
0     0% 14.34%       -2MB   0.7%  database/sql.(*DB).BeginTx.func1
0     0% 14.34%     1.02MB  0.36%  database/sql.(*DB).PingContext
0     0% 14.34%     1.02MB  0.36%  database/sql.(*DB).PingContext.func1
0     0% 14.34%       -2MB   0.7%  database/sql.(*DB).begin
0     0% 14.34%     0.50MB  0.17%  database/sql.(*DB).beginDC
0     0% 14.34%    -1.48MB  0.52%  database/sql.(*DB).conn
0     0% 14.34%    -0.50MB  0.17%  database/sql.(*DB).connectionOpener
0     0% 14.34%        1MB  0.35%  database/sql.(*DB).prepareDC.func2
0     0% 14.34%    -9.48MB  3.32%  database/sql.(*DB).retry
0     0% 14.34%    -8.50MB  2.97%  database/sql.(*Stmt).ExecContext
0     0% 14.34%    -8.50MB  2.97%  database/sql.(*Stmt).ExecContext.func1
0     0% 14.34%       -1MB  0.35%  database/sql.(*Stmt).connStmt
0     0% 14.34%     1.50MB  0.52%  database/sql.(*Tx).PrepareContext
0     0% 14.34%     0.50MB  0.17%  database/sql.ctxDriverPrepare
0     0% 14.34%    -3.50MB  1.22%  database/sql.ctxDriverStmtExec
0     0% 14.34%        1MB  0.35%  database/sql.withLock
0     0% 14.34%    17.56MB  6.14%  encoding/json.(*Decoder).Decode
0     0% 14.34%     1.01MB  0.35%  encoding/json.(*Decoder).readValue
0     0% 14.34%   -25.04MB  8.76%  encoding/json.(*Encoder).Encode
0     0% 14.34%    16.05MB  5.61%  encoding/json.(*decodeState).array
0     0% 14.34%     0.50MB  0.17%  encoding/json.(*decodeState).scanWhile
0     0% 14.34%    16.55MB  5.79%  encoding/json.(*decodeState).unmarshal
0     0% 14.34%    16.05MB  5.61%  encoding/json.(*decodeState).value
0     0% 14.34%    -2.50MB  0.88%  encoding/json.(*encodeState).marshal
0     0% 14.34%    -2.50MB  0.88%  encoding/json.(*encodeState).reflectValue
0     0% 14.34%    -1.50MB  0.53%  encoding/json.arrayEncoder.encode
0     0% 14.34%       -1MB  0.35%  encoding/json.indirect
0     0% 14.34%       -1MB  0.35%  encoding/json.mapEncoder.encode
0     0% 14.34%    -1.50MB  0.53%  encoding/json.sliceEncoder.encode
0     0% 14.34%     0.50MB  0.17%  encoding/json.stateBeginValue
0     0% 14.34%       -1MB  0.35%  encoding/json.stringEncoder
0     0% 14.34%    -1.50MB  0.53%  encoding/json.structEncoder.encode
0     0% 14.34%   -35.98MB 12.59%  github.com/MKhiriev/stunning-adventure/internal/handlers.(*Handler).BatchUpdateMetricJSON
0     0% 14.34%     0.50MB  0.17%  github.com/MKhiriev/stunning-adventure/internal/handlers.(*Handler).Init.NewRouter.NewMux.func4
0     0% 14.34%   -41.53MB 14.53%  github.com/MKhiriev/stunning-adventure/internal/handlers.(*Handler).WithLogging-fm.(*Handler).WithLogging.func1
0     0% 14.34%   -37.53MB 13.13%  github.com/MKhiriev/stunning-adventure/internal/handlers.GZip.func1
0     0% 14.34%  -270.83MB 94.74%  github.com/MKhiriev/stunning-adventure/internal/handlers.WithContext.func1
0     0% 14.34%    -0.50MB  0.17%  github.com/MKhiriev/stunning-adventure/internal/server.(*Server).ServerRun (inline)
0     0% 14.34%      -33MB 11.55%  github.com/MKhiriev/stunning-adventure/internal/service.(*DatabaseMetricsService).SaveAll
0     0% 14.34%      -33MB 11.55%  github.com/MKhiriev/stunning-adventure/internal/store.(*DB).SaveAll
0     0% 14.34%   -33.50MB 11.72%  github.com/MKhiriev/stunning-adventure/internal/store.(*DB).SaveAll.func1
0     0% 14.34%     0.50MB  0.17%  github.com/MKhiriev/stunning-adventure/internal/store.(*DB).checkIfRetryable
0     0% 14.34%      -33MB 11.55%  github.com/MKhiriev/stunning-adventure/internal/store.(*DB).withRetry
0     0% 14.34%     1.02MB  0.36%  github.com/MKhiriev/stunning-adventure/internal/store.NewConnectPostgres
0     0% 14.34%        2MB   0.7%  github.com/MKhiriev/stunning-adventure/internal/utils.Hash
0     0% 14.34%   -37.53MB 13.13%  github.com/go-chi/chi/v5.(*ChainHandler).ServeHTTP
0     0% 14.34%   -40.53MB 14.18%  github.com/go-chi/chi/v5.(*Mux).ServeHTTP
0     0% 14.34%   -37.53MB 13.13%  github.com/go-chi/chi/v5.(*Mux).routeHTTP
0     0% 14.34%   -41.53MB 14.53%  github.com/go-chi/chi/v5/middleware.Recoverer.func1
0     0% 14.34%     0.50MB  0.17%  github.com/jackc/pgx/v5.(*Conn).Deallocate
0     0% 14.34%    -7.50MB  2.62%  github.com/jackc/pgx/v5.(*Conn).Exec
0     0% 14.34%        1MB  0.35%  github.com/jackc/pgx/v5.(*Conn).Prepare
0     0% 14.34%    -7.50MB  2.62%  github.com/jackc/pgx/v5.(*Conn).exec
0     0% 14.34%    -7.50MB  2.62%  github.com/jackc/pgx/v5.(*Conn).execPrepared
0     0% 14.34%     0.50MB  0.17%  github.com/jackc/pgx/v5.(*ExtendedQueryBuilder).Build
0     0% 14.34%     0.50MB  0.17%  github.com/jackc/pgx/v5.(*ExtendedQueryBuilder).appendParam
0     0% 14.34%     0.50MB  0.17%  github.com/jackc/pgx/v5.(*ExtendedQueryBuilder).encodeExtendedParamValue
0     0% 14.34%       -1MB  0.35%  github.com/jackc/pgx/v5.ParseConfig (inline)
0     0% 14.34%       -1MB  0.35%  github.com/jackc/pgx/v5.ParseConfigWithOptions
0     0% 14.34%     0.50MB  0.17%  github.com/jackc/pgx/v5/pgconn.(*PgConn).Deallocate
0     0% 14.34%       -2MB   0.7%  github.com/jackc/pgx/v5/pgconn.(*PgConn).ExecPrepared
0     0% 14.34%        1MB  0.35%  github.com/jackc/pgx/v5/pgconn.(*PgConn).Prepare
0     0% 14.34%       -2MB   0.7%  github.com/jackc/pgx/v5/pgconn.(*PgConn).execExtendedPrefix
0     0% 14.34%    -2.50MB  0.87%  github.com/jackc/pgx/v5/pgconn.(*ResultReader).NextRow
0     0% 14.34%    -2.50MB  0.87%  github.com/jackc/pgx/v5/pgconn.(*ResultReader).receiveMessage
0     0% 14.34%    -0.50MB  0.17%  github.com/jackc/pgx/v5/pgconn.(*scramClient).recvServerFinalMessage
0     0% 14.34%       -1MB  0.35%  github.com/jackc/pgx/v5/pgconn.ConnectConfig
0     0% 14.34%       -1MB  0.35%  github.com/jackc/pgx/v5/pgconn.ParseConfigWithOptions
0     0% 14.34%    -0.50MB  0.17%  github.com/jackc/pgx/v5/pgconn.computeHMAC
0     0% 14.34%    -0.50MB  0.17%  github.com/jackc/pgx/v5/pgconn.computeServerSignature
0     0% 14.34%       -1MB  0.35%  github.com/jackc/pgx/v5/pgconn.connectOne
0     0% 14.34%       -1MB  0.35%  github.com/jackc/pgx/v5/pgconn.connectPreferred
0     0% 14.34%    -0.50MB  0.17%  github.com/jackc/pgx/v5/pgconn.defaultHost
0     0% 14.34%       -1MB  0.35%  github.com/jackc/pgx/v5/pgconn.defaultSettings
0     0% 14.34%     0.50MB  0.17%  github.com/jackc/pgx/v5/pgtype.(*Map).Encode
0     0% 14.34%     0.50MB  0.17%  github.com/jackc/pgx/v5/pgtype.(*derefPointerEncodePlan).Encode
0     0% 14.34%     1.02MB  0.36%  github.com/jackc/pgx/v5/pgtype.NewMap
0     0% 14.34%     1.02MB  0.36%  github.com/jackc/pgx/v5/pgtype.initDefaultMap
0     0% 14.34%    -3.50MB  1.22%  github.com/jackc/pgx/v5/stdlib.(*Stmt).ExecContext
0     0% 14.34%   -25.03MB  8.75%  github.com/rs/zerolog.(*Event).Any (partial-inline)
0     0% 14.34%   -30.54MB 10.68%  github.com/rs/zerolog.(*Event).Interface
0     0% 14.34%      -10MB  3.50%  github.com/rs/zerolog.(*Event).Msg (partial-inline)
0     0% 14.34%     0.50MB  0.17%  github.com/rs/zerolog.(*Event).Send
0     0% 14.34%    -9.50MB  3.32%  github.com/rs/zerolog.(*Event).msg
0     0% 14.34%    -9.50MB  3.32%  github.com/rs/zerolog.callerHook.Run
0     0% 14.34%   -30.54MB 10.68%  github.com/rs/zerolog.init.1.func1
0     0% 14.34%   -30.54MB 10.68%  github.com/rs/zerolog/internal/json.Encoder.AppendInterface
0     0% 14.34%    -0.50MB  0.17%  internal/poll.(*FD).Accept
0     0% 14.34%    -0.50MB  0.17%  internal/poll.accept
0     0% 14.34%     0.52MB  0.18%  main.main
0     0% 14.34%    -0.50MB  0.17%  net.(*TCPListener).Accept
0     0% 14.34%    -0.50MB  0.17%  net.(*TCPListener).accept
0     0% 14.34%    -0.50MB  0.17%  net.(*netFD).accept
0     0% 14.34%    -0.50MB  0.17%  net/http.(*Server).ListenAndServe
0     0% 14.34%    -0.50MB  0.17%  net/http.(*Server).Serve
0     0% 14.34%    -0.50MB  0.17%  net/http.(*connReader).backgroundRead
0     0% 14.34%   -41.53MB 14.53%  net/http.HandlerFunc.ServeHTTP
0     0% 14.34%     0.50MB  0.18%  net/http.newBufioReader
0     0% 14.34%   -40.53MB 14.18%  net/http.serverHandler.ServeHTTP
0     0% 14.34%       -1MB  0.35%  net/textproto.(*Reader).ReadMIMEHeader (inline)
0     0% 14.34%       -1MB  0.35%  os.Stat
0     0% 14.34%       -1MB  0.35%  reflect.(*MapIter).Value
0     0% 14.34%    15.55MB  5.44%  reflect.Value.Grow
0     0% 14.34%     0.50MB  0.17%  reflect.Value.Interface (inline)
0     0% 14.34%    15.55MB  5.44%  reflect.Value.grow
0     0% 14.34%     0.50MB  0.17%  reflect.valueInterface
0     0% 14.34%     0.50MB  0.17%  runtime.doInit (inline)
0     0% 14.34%     0.50MB  0.17%  runtime.doInit1
0     0% 14.34%     0.50MB  0.18%  runtime.findRunnable
0     0% 14.34%    -0.50MB  0.18%  runtime.handoffp
0     0% 14.34%     0.50MB  0.18%  runtime.injectglist
0     0% 14.34%     0.50MB  0.18%  runtime.injectglist.func1
0     0% 14.34%     1.02MB  0.36%  runtime.main
0     0% 14.34%     0.50MB  0.18%  runtime.mstart
0     0% 14.34%     0.50MB  0.18%  runtime.mstart0
0     0% 14.34%     0.50MB  0.18%  runtime.mstart1
0     0% 14.34%     0.50MB  0.18%  runtime.newm
0     0% 14.34%     0.50MB  0.18%  runtime.resetspinning
0     0% 14.34%    -0.50MB  0.18%  runtime.retake
0     0% 14.34%        1MB  0.35%  runtime.schedule
0     0% 14.34%     0.50MB  0.18%  runtime.startm
0     0% 14.34%    -0.50MB  0.18%  runtime.sysmon
0     0% 14.34%     0.50MB  0.18%  runtime.wakep
0     0% 14.34%     1.02MB  0.36%  sync.(*Once).Do (inline)
0     0% 14.34%     1.02MB  0.36%  sync.(*Once).doSlow
0     0% 14.34%     3.50MB  1.23%  sync.(*Pool).Get
0     0% 14.34%     0.50MB  0.18%  sync.(*Pool).Put
0     0% 14.34%     1.50MB  0.53%  sync.(*Pool).pin
0     0% 14.34%    -0.50MB  0.17%  syscall.Accept
0     0% 14.34%    -0.50MB  0.17%  time.AfterFunc
0     0% 14.34%     0.50MB  0.17%  time.init
```

## heap

```json
> go tool pprof -diff_base=base_heap.pprof 7_heap_db_conn.pprof

File: ___SERVER___DB___hashing
Type: inuse_space
Time: 2025-11-28 02:26:58 +05
Showing nodes accounting for -2563.49kB, 41.57% of 6166.94kB total
Dropped 5 nodes (cum <= 30.83kB)
flat  flat%   sum%        cum   cum%
-1028kB 16.67% 16.67%    -1028kB 16.67%  bufio.NewWriterSize (inline)
525.43kB  8.52%  8.15%   525.43kB  8.52%  github.com/jackc/pgx/v5/pgtype.(*Map).buildReflectTypeToType (inline)
-516.64kB  8.38% 16.53%  -516.64kB  8.38%  runtime.procresize
-516.01kB  8.37% 24.89%  -516.01kB  8.37%  github.com/jackc/pgx/v5/internal/iobufpool.init.0.func1
-516.01kB  8.37% 33.26%  -516.01kB  8.37%  io.init.func1
-512.22kB  8.31% 41.57%  -512.22kB  8.31%  runtime.malg
-512.05kB  8.30% 49.87%  -512.05kB  8.30%  context.(*cancelCtx).Done
512.02kB  8.30% 41.57%   512.02kB  8.30%  time.map.init.0
0     0% 41.57%  -516.01kB  8.37%  database/sql.(*DB).BeginTx
0     0% 41.57%  -516.01kB  8.37%  database/sql.(*DB).BeginTx.func1
0     0% 41.57%   525.43kB  8.52%  database/sql.(*DB).PingContext
0     0% 41.57%   525.43kB  8.52%  database/sql.(*DB).PingContext.func1
0     0% 41.57%  -516.01kB  8.37%  database/sql.(*DB).begin
0     0% 41.57%  -512.05kB  8.30%  database/sql.(*DB).connectionOpener
0     0% 41.57%  -516.01kB  8.37%  github.com/MKhiriev/stunning-adventure/internal/handlers.(*Handler).BatchUpdateMetricJSON
0     0% 41.57%  -516.01kB  8.37%  github.com/MKhiriev/stunning-adventure/internal/handlers.(*Handler).WithHashing-fm.(*Handler).WithHashing.func1
0     0% 41.57%  -516.01kB  8.37%  github.com/MKhiriev/stunning-adventure/internal/handlers.(*Handler).WithLogging-fm.(*Handler).WithLogging.func1
0     0% 41.57%  -516.01kB  8.37%  github.com/MKhiriev/stunning-adventure/internal/handlers.GZip.func1
0     0% 41.57%  -516.01kB  8.37%  github.com/MKhiriev/stunning-adventure/internal/handlers.WithContext.func1
0     0% 41.57%  -516.01kB  8.37%  github.com/MKhiriev/stunning-adventure/internal/service.(*DatabaseMetricsService).SaveAll
0     0% 41.57%  -516.01kB  8.37%  github.com/MKhiriev/stunning-adventure/internal/service.(*ValidatingMetricsService).SaveAll
0     0% 41.57%  -516.01kB  8.37%  github.com/MKhiriev/stunning-adventure/internal/store.(*DB).SaveAll
0     0% 41.57%  -516.01kB  8.37%  github.com/MKhiriev/stunning-adventure/internal/store.(*DB).SaveAll.func1
0     0% 41.57%  -516.01kB  8.37%  github.com/MKhiriev/stunning-adventure/internal/store.(*DB).saveAllMetrics
0     0% 41.57%  -516.01kB  8.37%  github.com/MKhiriev/stunning-adventure/internal/store.(*DB).withRetry
0     0% 41.57%   525.43kB  8.52%  github.com/MKhiriev/stunning-adventure/internal/store.NewConnectPostgres
0     0% 41.57%  -516.01kB  8.37%  github.com/go-chi/chi/v5.(*ChainHandler).ServeHTTP
0     0% 41.57%  -516.01kB  8.37%  github.com/go-chi/chi/v5.(*Mux).ServeHTTP
0     0% 41.57%  -516.01kB  8.37%  github.com/go-chi/chi/v5.(*Mux).routeHTTP
0     0% 41.57%  -516.01kB  8.37%  github.com/go-chi/chi/v5/middleware.Recoverer.func1
0     0% 41.57%  -516.01kB  8.37%  github.com/jackc/pgx/v5/internal/iobufpool.Get
0     0% 41.57%  -516.01kB  8.37%  github.com/jackc/pgx/v5/pgconn.ConnectConfig
0     0% 41.57%  -516.01kB  8.37%  github.com/jackc/pgx/v5/pgconn.ParseConfigWithOptions.func1
0     0% 41.57%  -516.01kB  8.37%  github.com/jackc/pgx/v5/pgconn.connectOne
0     0% 41.57%  -516.01kB  8.37%  github.com/jackc/pgx/v5/pgconn.connectPreferred
0     0% 41.57%  -516.01kB  8.37%  github.com/jackc/pgx/v5/pgproto3.NewFrontend
0     0% 41.57%  -516.01kB  8.37%  github.com/jackc/pgx/v5/pgproto3.newChunkReader (inline)
0     0% 41.57%   525.43kB  8.52%  github.com/jackc/pgx/v5/pgtype.NewMap
0     0% 41.57%   525.43kB  8.52%  github.com/jackc/pgx/v5/pgtype.initDefaultMap
0     0% 41.57%  -516.01kB  8.37%  io.Copy (inline)
0     0% 41.57%  -516.01kB  8.37%  io.CopyN
0     0% 41.57%  -516.01kB  8.37%  io.copyBuffer
0     0% 41.57%  -516.01kB  8.37%  io.discard.ReadFrom
0     0% 41.57%   525.43kB  8.52%  main.main
0     0% 41.57%  -516.01kB  8.37%  net/http.(*chunkWriter).close
0     0% 41.57%  -516.01kB  8.37%  net/http.(*chunkWriter).writeHeader
0     0% 41.57% -2060.02kB 33.40%  net/http.(*conn).serve
0     0% 41.57%  -516.01kB  8.37%  net/http.(*response).finishRequest
0     0% 41.57%  -516.01kB  8.37%  net/http.HandlerFunc.ServeHTTP
0     0% 41.57%    -1028kB 16.67%  net/http.newBufioWriterSize
0     0% 41.57%  -516.01kB  8.37%  net/http.serverHandler.ServeHTTP
0     0% 41.57%   512.02kB  8.30%  runtime.doInit (inline)
0     0% 41.57%   512.02kB  8.30%  runtime.doInit1
0     0% 41.57%     -513kB  8.32%  runtime.findRunnable
0     0% 41.57%     -513kB  8.32%  runtime.injectglist
0     0% 41.57%     -513kB  8.32%  runtime.injectglist.func1
0     0% 41.57%  1037.45kB 16.82%  runtime.main
0     0% 41.57%     -513kB  8.32%  runtime.mcall
0     0% 41.57%      513kB  8.32%  runtime.mstart
0     0% 41.57%      513kB  8.32%  runtime.mstart0
0     0% 41.57%      513kB  8.32%  runtime.mstart1
0     0% 41.57%  -512.22kB  8.31%  runtime.newproc.func1
0     0% 41.57%  -512.22kB  8.31%  runtime.newproc1
0     0% 41.57%     -513kB  8.32%  runtime.park_m
0     0% 41.57%      513kB  8.32%  runtime.resetspinning
0     0% 41.57%  -516.64kB  8.38%  runtime.rt0_go
0     0% 41.57%  -516.64kB  8.38%  runtime.schedinit
0     0% 41.57%  -512.22kB  8.31%  runtime.systemstack
0     0% 41.57%      513kB  8.32%  runtime.wakep
0     0% 41.57%   525.43kB  8.52%  sync.(*Once).Do (inline)
0     0% 41.57%   525.43kB  8.52%  sync.(*Once).doSlow
0     0% 41.57% -1032.02kB 16.73%  sync.(*Pool).Get
0     0% 41.57%   512.02kB  8.30%  time.init
```
