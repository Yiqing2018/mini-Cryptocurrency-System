import matplotlib.pyplot as plt
from datetime import datetime
import fileinput
import re
from time import strptime
import glob





# find out all transactions, add into a list
bandwidth=[]
input = open("all_log.log", 'r')
for line in input:
    line = line.split()
    if 'bandwidth' in line:
        bw=(int)(line[3])
        bandwidth.append(bw)
input.close()


loc=range(0, len(bandwidth))
fig, ax = plt.subplots( nrows=1, ncols=1 )  # create figure & 1 axis
ax.scatter(loc, bandwidth,s=3**3, marker="s", color="red")
fig.suptitle('Bandwitdh - how many bytes are sent/received')
plt.xlabel('every node')
plt.xticks([])
plt.ylabel('bytes')
fig.savefig('Bandwitdh.png')   # save the figure to file
plt.close(fig) 

# plt.scatter(loc,bandwidth)
# plt.show()



