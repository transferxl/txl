USER=user
PASSWORD=password
LOG=performance.csv
# 1 GB
# dd if=/dev/zero bs=1M count=1024 2> /dev/null | ./txl put -l $LOG -u $USER -p $PASSWORD | ./txl get -l $LOG -o /dev/null
# 10 GB
# dd if=/dev/zero bs=1M count=10240 2> /dev/null | ./txl put -l $LOG -u $USER -p $PASSWORD | ./txl get -l $LOG -o /dev/null
# 100 GB
# dd if=/dev/zero bs=1M count=102400 2> /dev/null | ./txl put -l $LOG -u $USER -p $PASSWORD | ./txl get -l $LOG -o /dev/null
# 200 GB
# dd if=/dev/zero bs=1M count=204800 2> /dev/null | ./txl put -l $LOG -u $USER -p $PASSWORD | ./txl get -l $LOG -o /dev/null
# 300 GB
# dd if=/dev/zero bs=1M count=307200 2> /dev/null | ./txl put -l $LOG -u $USER -p $PASSWORD | ./txl get -l $LOG -o /dev/null
