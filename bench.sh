#! /usr/bin/env bash
# bench.sh runs the standard benchmark suite for comparing performance between
# different versions of the solver. Output should be redirected to a file, and
# a copy of the most recent benchmark results should be saved in the `bench`
# directory for comparison. After generating a new results file, use `benchstat`
# to compare the results with a previous run.

# Benchmarks are divided into two groups, "parse" and "solve", and the groups
# to run are controlled by the BENCH_TESTS variable.  The default value is
# "solve" because parsing is not particularly time-critical, and it is rarely
# necessary to spend time re-running those tests.  Set "BENCH_TESTS=all" to
# run both groups of tests.

# By default, each benchmark is run 10 times to ensure consistent results, and
# all runs that will be captured for future comparisons should use the default.
# However, a complete comparison run takes a considerable amount of time.
# During initial testing, you can override the repeat count by setting the
# BENCH_COUNT environment variable to reduce the duration.

BENCH_TESTS=${BENCH_TESTS:-solve}
BENCH_COUNT=${BENCH_COUNT:-10}

if [ "$BENCH_TESTS" = "parse" ] || [ "$BENCH_TESTS" = "all" ]; then
    # Run parsing initialization benchmarks. These are quick so there is no need
    # to run them for extra time.
    go test -bench=BenchmarkParseString -benchmem --count=$BENCH_COUNT ./internal/puzzle
    go test -bench=BenchmarkNewSolver -benchmem --count=$BENCH_COUNT ./internal/solver
fi

if [ "$BENCH_TESTS" = "solve" ] || [ "$BENCH_TESTS" = "all" ]; then
    # Run the Solve benchmark for comparing each test puzzle separately.
    # The slowest of the test cases should still run in about 1ms, so a default
    # 1s benchmark run should still give us around a thousand iterations.
    go test -bench='BenchmarkSolve$' --count=$BENCH_COUNT ./internal/solver

    # Run the comparison subset for a longer duration to get better memory usage
    # statistics.
    go test -bench=BenchmarkComparison -benchmem -benchtime=5s --count=$BENCH_COUNT ./internal/solver
fi
