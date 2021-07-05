#!/bin/bash
i=0
N=50
header="Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2MjU0OTQwMjcsImlkIjoxLCJ1c2VybmFtZSI6InVzZXIxIn0.QkdyCUkXMZDSumwEbdAd44syfOwH0VzWRNS2yp3oXIaQRSNT3wIEHotod1m_w5EKoOv1ocQJs5kOxHvtKAqD4Q"
addr="127.0.0.1:80/gateway/playback"

while [ $i -lt $N ]; do
	d=$(jq -n --arg id "hello-$i" '{ID:$id}')
	curl $addr -d "$d" -H "$header"
	let "i++"
done
