cd profiles

go tool pprof -output=base_cpu.pprof -seconds=25 localhost:8081/debug/pprof/profile
go tool pprof -output=base_allocs.pprof localhost:8081/debug/pprof/allocs
go tool pprof -output=base_heap.pprof localhost:8081/debug/pprof/heap

# first optimisation
go tool pprof -output=1_cpu_WithContextDeleted.pprof -seconds=25 localhost:8081/debug/pprof/profile
go tool pprof -output=1_allocs_WithContextDeleted.pprof localhost:8081/debug/pprof/allocs
go tool pprof -output=1_heap_WithContextDeleted.pprof localhost:8081/debug/pprof/heap

go tool pprof -top -diff_base=base_cpu.pprof 1_cpu_WithContextDeleted.pprof
go tool pprof -top -diff_base=base_allocs.pprof 1_allocs_WithContextDeleted.pprof
go tool pprof -top -diff_base=base_heap.pprof 1_heap_WithContextDeleted.pprof

# second optimisation
go tool pprof -output=2_cpu_logging_middleware_fix.pprof -seconds=25 localhost:8081/debug/pprof/profile
go tool pprof -output=2_allocs_logging_middleware_fix.pprof localhost:8081/debug/pprof/allocs
go tool pprof -output=2_heap_logging_middleware_fix.pprof localhost:8081/debug/pprof/heap

go tool pprof -top -diff_base=1_cpu_WithContextDeleted.pprof 2_cpu_logging_middleware_fix.pprof
go tool pprof -top -diff_base=1_allocs_WithContextDeleted.pprof 2_allocs_logging_middleware_fix.pprof
go tool pprof -top -diff_base=1_heap_WithContextDeleted.pprof 2_heap_logging_middleware_fix.pprof

# third optimisation
go tool pprof -output=3_cpu_gzip_sync_pool.pprof -seconds=25 localhost:8081/debug/pprof/profile
go tool pprof -output=3_allocs_gzip_sync_pool.pprof localhost:8081/debug/pprof/allocs
go tool pprof -output=3_heap_gzip_sync_pool.pprof localhost:8081/debug/pprof/heap

go tool pprof -top -diff_base=2_cpu_logging_middleware_fix.pprof 3_cpu_gzip_sync_pool.pprof
go tool pprof -top -diff_base=2_allocs_logging_middleware_fix.pprof 3_allocs_gzip_sync_pool.pprof
go tool pprof -top -diff_base=2_heap_logging_middleware_fix.pprof 3_heap_gzip_sync_pool.pprof

# fourth optimisation
go tool pprof -output=4_cpu_hasher.pprof -seconds=25 localhost:8081/debug/pprof/profile
go tool pprof -output=4_allocs_hasher.pprof localhost:8081/debug/pprof/allocs
go tool pprof -output=4_heap_hasher.pprof localhost:8081/debug/pprof/heap

go tool pprof -top -diff_base=3_cpu_gzip_sync_pool.pprof 4_cpu_hasher.pprof
go tool pprof -top -diff_base=3_allocs_gzip_sync_pool.pprof 4_allocs_hasher.pprof
go tool pprof -top -diff_base=3_heap_gzip_sync_pool.pprof 4_heap_hasher.pprof

# fifth optimisation - maybe delete
go tool pprof -output=5_cpu_hasher.pprof -seconds=25 localhost:8081/debug/pprof/profile
go tool pprof -output=5_allocs_hasher.pprof localhost:8081/debug/pprof/allocs
go tool pprof -output=5_heap_hasher.pprof localhost:8081/debug/pprof/heap


go tool pprof -top -diff_base=base_cpu.pprof 5_cpu_hasher.pprof
go tool pprof -top -diff_base=base_allocs.pprof 5_allocs_hasher.pprof
go tool pprof -top -diff_base=base_heap.pprof 5_heap_hasher.pprof


# sixth optimisation - logger fix + max db
go tool pprof -output=6_cpu_logger.pprof -seconds=25 localhost:8081/debug/pprof/profile
go tool pprof -output=6_allocs_logger.pprof localhost:8081/debug/pprof/allocs
go tool pprof -output=6_heap_logger.pprof localhost:8081/debug/pprof/heap

go tool pprof -top -diff_base=5_cpu_hasher.pprof 6_cpu_logger.pprof
go tool pprof -top -diff_base=5_allocs_hasher.pprof 6_allocs_logger.pprof
go tool pprof -top -diff_base=5_heap_hasher.pprof 6_heap_logger.pprof

# seventh optimisation - max db connections
go tool pprof -output=7_cpu_db_conn.pprof -seconds=25 localhost:8081/debug/pprof/profile
go tool pprof -output=7_allocs_db_conn.pprof localhost:8081/debug/pprof/allocs
go tool pprof -output=7_heap_db_conn.pprof localhost:8081/debug/pprof/heap

go tool pprof -top -diff_base=6_cpu_logger.pprof 7_cpu_db_conn.pprof
go tool pprof -top -diff_base=6_allocs_logger.pprof 7_allocs_db_conn.pprof
go tool pprof -top -diff_base=6_heap_logger.pprof 7_heap_db_conn.pprof