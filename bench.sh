# bench.sh runs the standard benchmark suite for comparing performance between
# different versions of the solver. Output should be redirected to a file, and
# a copy of the most recent benchmark results should be saved in the `bench`
# directory for comparison. After generating a new results file, use `benchstat`
# to compare the results with a previous run.

# By default, each benchmark is run 10 times to ensure consistent results, and
# all runs that will be captured for future comparisons should use the default.
# However, a complete comparison run takes a considerable amount of time.
# During initial testing, you can override the repeat count by setting the
# BENCH_COUNT environment variable to reduce the duration.

BENCH_COUNT=${BENCH_COUNT:-10}

# Run parsing initialization benchmarks. These are quick so there is no need
# to run them for extra time.
go test -bench=BenchmarkParseString -benchmem --count=$BENCH_COUNT ./internal/puzzle
go test -bench=BenchmarkNewSolver -benchmem --count=$BENCH_COUNT ./internal/solver

# Run the Solve benchmark for comparing each test puzzle separately.
# The slowest of the test cases should still run in about 1ms, so a default 1s
# benchmark run should still give us around a thousand iterations.
go test -bench='BenchmarkSolve$' --count=$BENCH_COUNT ./internal/solver

# Run the comparison subset for a longer duration to get better memory usage
# statistics.
go test -bench=BenchmarkComparison -benchmem -benchtime=5s --count=$BENCH_COUNT ./internal/solver
