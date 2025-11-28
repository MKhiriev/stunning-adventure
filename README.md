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
Entering interactive mode (type "help" for commands, "o" for options)
(pprof) top
Showing nodes accounting for -32.11MB, 11.23% of 285.87MB total
Dropped 3 nodes (cum <= 1.43MB)
Showing top 10 nodes out of 211
      flat  flat%   sum%        cum   cum%
  -18.54MB  6.48%  6.48%   -18.54MB  6.48%  bytes.growSlice
    7.02MB  2.45%  4.03%     7.02MB  2.45%  reflect.growslice
      -7MB  2.45%  6.48%   -17.50MB  6.12%  github.com/MKhiriev/stunning-adventure/internal/store.(*DB).saveAllMetrics
   -5.50MB  1.92%  8.40%   -30.54MB 10.68%  github.com/rs/zerolog.init.func7
      -5MB  1.75% 10.15%   -23.54MB  8.23%  bytes.(*Buffer).grow
    4.01MB  1.40%  8.75%     4.01MB  1.40%  io.ReadAll
      -4MB  1.40% 10.15%       -7MB  2.45%  github.com/jackc/pgx/v5/pgconn.(*ResultReader).Read
   -3.61MB  1.26% 11.41%    -3.61MB  1.26%  compress/flate.(*dictDecoder).init
    3.51MB  1.23% 10.18%     3.51MB  1.23%  encoding/json.(*Decoder).refill
      -3MB  1.05% 11.23%       -3MB  1.05%  github.com/jackc/pgx/v5/pgconn.(*PgConn).makeCommandTag (inline)
```

## heap
```json
> go tool pprof -diff_base=base_heap.pprof 7_heap_db_conn.pprof

File: ___SERVER___DB___hashing
Type: inuse_space
Time: 2025-11-28 02:26:58 +05
Entering interactive mode (type "help" for commands, "o" for options)
(pprof) top
Showing nodes accounting for -2063.60kB, 33.46% of 6166.94kB total
Dropped 29 nodes (cum <= 30.83kB)
Showing top 10 nodes out of 45
      flat  flat%   sum%        cum   cum%
 -516.64kB  8.38%  8.38%  -516.64kB  8.38%  runtime.procresize
 -516.01kB  8.37% 16.74%  -516.01kB  8.37%  github.com/jackc/pgx/v5/internal/iobufpool.init.0.func1
 -516.01kB  8.37% 25.11%  -516.01kB  8.37%  io.init.func1
    -514kB  8.33% 33.45%     -514kB  8.33%  bufio.NewWriterSize
    -514kB  8.33% 41.78%     -514kB  8.33%  bufio.NewReaderSize
     513kB  8.32% 33.46%      513kB  8.32%  runtime.allocm
  512.28kB  8.31% 25.16%   512.28kB  8.31%  crypto/x509/pkix.map.init.0
 -512.22kB  8.31% 33.46%  -512.22kB  8.31%  runtime.malg
 -512.05kB  8.30% 41.77%  -512.05kB  8.30%  context.(*cancelCtx).Done
  512.05kB  8.30% 33.46%  1024.34kB 16.61%  runtime.main
```
