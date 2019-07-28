import matplotlib.pyplot as plt
from datetime import datetime
import fileinput
import re
from time import strptime
import glob

# get all log files'names under the directory
f_names=[]
for filename in glob.glob('*.log'):
    f_names.append(filename)
f= open("all_log.log","w+")

# merge all log files and sorted with timestamp
# write into all_log.log
lines = list(fileinput.input(f_names))
t_fmt = '%Y/%m/%d %H:%M:%S' # format of time stamps
t_pat = re.compile(r'(\d+/\d+/\d+ \d+:\d+:\d+)') # pattern to extract timestamp
for l in sorted(lines, key=lambda l: strptime(t_pat.search(l).group(1), t_fmt)):
    f.write(l)
f.close()

# find out all transactions, add into a list
all_block= []
start_time={}
end_time={}
input = open("all_log.log", 'r')
for line in input:
    line = line.split()
    if 'CreateInitialBlock' in line:
        blockId=line[3]
        time = line[0]+" "+line[1]
        all_block.append(blockId)
        start_time[blockId]=time
    if 'ReceiveVerifiedBlock' in line:
        blockId=line[3]
        time = line[0]+" "+line[1]
        if blockId not in start_time:
            all_block.append(blockId)
            start_time[blockId]=time
        else:
            end_time[line[3]]=time

input.close()


propagation_time=[]

for tran in all_block:
    if tran not in start_time  or tran not in end_time:
        continue
    start=datetime.strptime(start_time[tran], '%Y/%m/%d %H:%M:%S')
    end=datetime.strptime(end_time[tran], '%Y/%m/%d %H:%M:%S')
    delta=(end-start).total_seconds()
    if delta!=0:
    	propagation_time.append(delta)

loc=range(0, len(propagation_time))
fig, ax = plt.subplots( nrows=1, ncols=1 )  # create figure & 1 axis
ax.scatter(loc, propagation_time,s=2**2, marker="s")
plt.xticks([])
fig.suptitle('Block Propagation Time')
plt.xlabel('every block')
plt.ylabel('propagation time')
fig.savefig('BlockPropagationTime.png')   # save the figure to file
plt.close(fig) 



