# /bin/bash
CLUSTER_ROOT_PATH="/etc/backup"
FULL_FILE_NAME="backup"
MYSQL_DATA_DIR="/var/lib/mysql"
gunzip "$CLUSTER_ROOT_PATH/$FULL_FILE_NAME"
FULL_FILE_NAME_A=$(echo "$FULL_FILE_NAME" | sed -e "s/.gz//g")
xbstream -x -C "$MYSQL_DATA_DIR" <"$CLUSTER_ROOT_PATH/$FULL_FILE_NAME_A"
BACKUP_TIME=$(echo $FULL_FILE_NAME | sed -e "s/$CLUSTER_MARK//g" | cut -d "_" -f 2 | cut -d '.' -f 1)
xtrabackup --decompress --parallel=6 --compress-threads=6 --target-dir="$MYSQL_DATA_DIR"
xtrabackup --defaults-file=/etc/my_$MYSQL_PORT.cnf --prepare --apply-log-only --target-dir=$MYSQL_DATA_DIR
xtrabackup --defaults-file=/etc/my_$MYSQL_PORT.cnf --prepare --apply-log-only --target-dir=$MYSQL_DATA_DIR
xtrabackup --decompress --parallel=6 --compress-threads=6 --target-dir="$MYSQL_DATA_DIR"

