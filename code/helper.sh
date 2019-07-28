#!/bin/bash
dir="/home/MPs/mp2/demo"
cd $dir
tmp="./demo node"
for i in {1..20}
do
	port=`expr $i + 5250`
	com="$tmp""$i "" $port"" &"
	eval $com
done
