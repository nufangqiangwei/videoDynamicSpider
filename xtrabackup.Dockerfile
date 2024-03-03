FROM oraclelinux:8.9

COPY percona-xtrabackup-8.0.35-30-Linux-x86_64.glibc2.17.tar.gz /home
RUN chmod 644 /home/percona-xtrabackup-8.0.35-30-Linux-x86_64.glibc2.17.tar.gz
RUN tar -zxvf /home/percona-xtrabackup-8.0.35-30-Linux-x86_64.glibc2.17.tar.gz
RUN mkdir /usr/local/xtrabackup
RUN mv /home/percona-xtrabackup-8.0.35-30-Linux-x86_64.glibc2.17/* /usr/local/xtrabackup/
RUN ln -sf /usr/local/xtrabackup/bin/* /usr/bin/