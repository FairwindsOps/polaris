#! /bin/bash

go tool cover -func=./test/cover.out > ./test/coverage-report.txt

total=0
count=0

while read l; do
  fields=($l)
  fn_coverage=${fields[2]::-3}
  total=$((total + fn_coverage))
  count=$((count + 1))
done <./test/coverage-report.txt

coverage=$(( total / count ))
echo $coverage
color="239922"
if [[ $coverage -lt 50 ]] ; then
  color="C21807"
elif [[ $coverage -lt 70 ]] ; then
  color="FFD300"
fi
sed -i -- "s/^\[coverage-image\].*$/\[coverage-image\]: https:\/\/img.shields.io\/static\/v1.svg?label=Coverage\&message=$coverage%25\&color=$color/" README.md
