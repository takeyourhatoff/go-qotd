#!/bin/bash

for i in {1..20}; do
	curl -s -X GET "http://api.forismatic.com/api/1.0/?method=getQuote&format=text&lang=en&key="$i >> quotes.csv;
	echo "" >> quotes.csv;
done