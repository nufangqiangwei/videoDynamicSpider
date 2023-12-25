# /bin/bash
function loading_settings()
{
    [ -f /etc/init.d/functions ] && source /etc/init.d/functions
}
user="root"
passwd="p0o9i8u7"
timestamp_name="`date +'%Y%m%d%H%M%S'`"
logging="/backup/xtrabackup_log_${timestamp_name}.txt"

full_begin_time=$(date -d "today" +%s)
# flush binlog before full dump
action "`date '+%Y-%m-%d %T'` [INFO] Begining full backup, please wait." /bin/true

cur_binlog_name=`mysql -u ${user} --password=${passwd} -e "SHOW MASTER STATUS;" -N | awk '{print $1}'`
mysqladmin flush-logs -u ${user} --password=${passwd}
log_end_time=$(date -d "today" +%s)
let elapsed=log_end_time-full_begin_time

tmpfile=/etc/backup/dump_$RANDOM.sh
echo "#!/bin/sh" > ${tmpfile}
echo "xtrabackup  --user=${user} --password=${passwd} --backup --stream=xbstream --compress | gzip -2  > /backup/${timestamp_name}.xbstream.gz " >> ${tmpfile}

chmod 700 ${tmpfile}
/bin/bash ${tmpfile} > ${logging} 2>&1
[ `tail -n 1 ${logging} | grep -i 'completed OK' | wc -l` -ge 1 ] && {
  action "`date '+%Y-%m-%d %T'` [INFO] Full backup has been finished successfully!" /bin/true
  [ -f ${tmpfile} ] && rm -rf $tmpfile
  action "`date '+%Y-%m-%d %T'` [INFO] Finished Full DB dump in $elapsed Seconds!" /bin/true
} || {
  action "`date '+%Y-%m-%d %T'` [ERROR] Failed to do full backup!" /bin/false
  [ -f ${tmpfile} ] && rm -rf ${tmpfile}
  exit 1
}
