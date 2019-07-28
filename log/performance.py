import matplotlib.pyplot as plt
from datetime import datetime
import fileinput
import re
from time import strptime
import re


# find out all transactions, add into a list

accounts={}
input = open("all_log.log", 'r')
for line in input:
    line = line.split()
    if 'RealBalance' in line:
        for i in range(len(line)):
            s=line[i]
            if s.startswith('0:'):
                money=(int)(s[2:14])
                print(money)
                if money in accounts:
                    counts=accounts[money]
                    accounts[money]=counts+1
                else:
                    accounts[money]=1
input.close()


lists = sorted(accounts.items())
x, y = zip(*lists)

fig, ax = plt.subplots( nrows=1, ncols=1 )  # create figure & 1 axis
ax.scatter(x, y,s=3**3, marker="s")
fig.suptitle('agreement of balance')
plt.xlabel('balance')
plt.xticks([])
plt.ylabel('count')
fig.savefig('performance.png')   # save the figure to file
plt.close(fig) 



