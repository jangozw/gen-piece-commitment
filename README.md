# gen-piece-commitment

用来辅助lotus-miner接收deal的验证环节, 将lotus-miner import-data 中的生产piece cid 独立部署,减轻miner负载

将本程序部署在任意机器，只要能连接配置文件里的数据库等信息即可。lotus-miner根据数据库记录的结果进行下一步工作

## 编译

```shell
make build
```

生成的编译文件： gen-piece-commitment

## 运行

准备config.toml 配置的是数据库信息
```toml
[mysql]
host="127.0.0.1:3306"
user="root"
password="123456"
dbname="deal"

[redis]
host="127.0.0.1:6379"
password=""
db=0
poolsize=100

[log]
levels=[{ name = "inited", level = "debug" },{ name = "handler", level = "debug" }]
```


### 导入原始交易数据

```shell
./gen-piece-commitment import --c config.toml --file import.txt --miner t01000
```
import.txt
```text
sdfasdfsdfsdfasd  /var/data/a.car
12fasdfsdfsdfasd  /var/data/a.car
```
或:

```shell
./gen-piece-commitment import --c config.toml --file import.txt 
```
import.txt
```text
sdfasdfsdfsdfasd  /var/data/a.car f01233
12fasdfsdfsdfasd  /var/data/a.car f03456
```

导入deal成功日志：

```text
2023-02-28T18:24:11.792+0800	INFO	handler	handler/piecehandler.go:118	import deal line 1 success. proposal_cid: xxaddxxllllllll
2023-02-28T18:24:11.813+0800	INFO	handler	handler/piecehandler.go:118	import deal line 2 success. proposal_cid: 55xdddxaddxxllllllll
2023-02-28T18:24:11.813+0800	INFO	handler	handler/piecehandler.go:118	import deal line 0 success. proposal_cid: xxxxllllllll
2023-02-28T18:24:11.813+0800	INFO	handler	handler/piecehandler.go:118	import deal line 3 success. proposal_cid: 55333xdddxaddxxllllllll
2023-02-28T18:24:11.813+0800	INFO	handler	handler/piecehandler.go:122	Import deal by import.txt end. total: 4 succ: 4
```

### 运行批量生成piece cid


```shell
./gen-piece-commitment run --c config.toml
```
