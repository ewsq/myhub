# MyHub简介

MyHub是一个由Go开发高性能MySQL Proxy项目，MyHub在满足基本的读写分离的功能上，致力于简化MySQL分库分表操作。
MyHub和其它数据库中间件相比最大特点是做到最大限度的仿真MySql。

## 主要功能：

### 基础功能

1, 支持SQL读写分离。
2, 支持透明的MySQL连接池，不必每次新建连接。
3, 支持多个slave，slave之间通过权值进行负载均衡。
4, 支持读写分离。
5, 支持多租户。
6, 支持主流语言（java,php,python,C/C++,Go)SDK的mysql的prepare特性。
7, 支持到后端DB的最大连接数限制。
8, 支持SQL日志及慢日志输出。
9, 支持客户端IP访问白名单机制，只有白名单中的IP才能访问MyHub。
10, 支持字符集设置。
11, 支持last_insert_id功能。
12, 支持show databases,show tables

### 分片功能

1, 支持按整数的hash和range分表方式。
2, 支持按年、月、日维度的时间分表方式。
3, 支持跨节点分表，子表可以分布在不同的节点。
4, 支持跨节点的count,sum,max和min等聚合函数。
5, 支持单个分表的join操作，即支持分表和另一张不分表的join操作。
6, 支持跨节点的order by,group by,limit等操作。
7, 支持将sql发送到特定的节点执行。
8, 支持事务。
9, 支持数据库直接代理转发。
10, 支持（insert,delete,update,replace）到多个node上的子表。
11, 支持自动在多个node上创建分表。
12, 支持主键自增长ID。

## License

MyHub采用Apache 2.0协议.