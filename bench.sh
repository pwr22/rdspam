#!/bin/bash

cmd="rdspam"
size="1G"

bench_devnull_direct="./$cmd -s $size > /dev/null"
bench_devnull_pv="./$cmd -s $size | pv > /dev/null"
bench_devnull_cat="./$cmd -s $size | cat - > /dev/null"
bench_devnull_cat_subsh="cat <(./$cmd -s $size) > /dev/null"

go build -o $cmd

echo "/dev/null benchmarks"
echo

echo $bench_devnull_direct
eval $bench_devnull_direct

echo

echo $bench_devnull_pv
eval $bench_devnull_pv

echo

echo $bench_devnull_cat
eval $bench_devnull_cat

echo

echo $bench_devnull_cat_subsh
eval $bench_devnull_cat_subsh

tmpfs_file=$(mktemp /tmp/bench-rdspam.XXXXXX)

bench_tmpfs_file_direct="./$cmd -s $size > $tmpfs_file"
bench_tmpfs_file_pv="./$cmd -s $size | pv > $tmpfs_file"
bench_tmpfs_file_cat="./$cmd -s $size | cat - > $tmpfs_file"

echo
echo "file benchmarks"
echo

echo $bench_tmpfs_file_direct
eval $bench_tmpfs_file_direct

echo

echo $bench_tmpfs_file_pv
eval $bench_tmpfs_file_pv

echo

echo $bench_tmpfs_file_cat
eval $bench_tmpfs_file_cat

rm $tmpfs_file

reg_file=$(mktemp bench-rdspam.XXXXXX)

bench_reg_file_direct="./$cmd -s $size > $reg_file"
bench_reg_file_pv="./$cmd -s $size | pv > $reg_file"
bench_reg_file_cat="./$cmd -s $size | cat - > $reg_file"

echo
echo "file benchmarks"
echo

echo $bench_reg_file_direct
eval $bench_reg_file_direct

echo

echo $bench_reg_file_pv
eval $bench_reg_file_pv

echo

echo $bench_reg_file_cat
eval $bench_reg_file_cat

rm $reg_file