#!/usr/bin/env bash

sudo mariabackup --host=127.0.0.1 --port=3306 --user=root --password= --backup --version-check --datadir=/var/lib/mysql --target_dir=/backup/mariabackup_manual/ --extra-lsndir=/backup/mariabackup_manual/ --parallel=6 --stream=xbstream  > backup.xb
#sudo mariabackup --host=127.0.0.1 --port=3306 --user=root --password= --backup --version-check --datadir=/var/lib/mysql --target_dir=/backup/mariabackup_manual/ --extra-lsndir=/backup/mariabackup_manual/ --parallel=6 --stream=xbstream  | pigz -p 12 --fast > backup.x
