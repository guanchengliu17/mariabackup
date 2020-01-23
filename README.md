# Golang Wrapper For Mariabackup

Mariabackup is an open source tool provided by MariaDB for performing physical online backups of InnoDB, Aria and MyISAM tables. For InnoDB, “hot online” backups are possible. It was originally forked from Percona XtraBackup 2.3.8. It is available on Linux and Windows.

## Dependencies

Install mariabackup:
```
$ sudo apt-get install software-properties-common
$ sudo apt-key adv --fetch-keys 'https://mariadb.org/mariadb_release_signing_key.asc'
$ sudo add-apt-repository 'deb [arch=amd64,arm64,ppc64el] http://mirrors.tuna.tsinghua.edu.cn/mariadb/repo/10.4/ubuntu bionic main'
$ sudo apt-get install mariadb-backup
```

## Build

```
$ go build -o mariabackup-wrapper
```

## Usage

Full backup:
```
$ ./mariabackup-wrapper backup -username=root -mode=full 
```

Incremental backup:
```
$ ./mariabackup-wrapper backup -username=root -mode=incremental
```

Restore backup:
```
$ ./mariabackup-wrapper restore
```
