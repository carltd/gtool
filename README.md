# gtool

gtool is a development command-line tool for carltd

## Install

    go get github.com/carltd/gtool
    go install
    $GOPATH/bin/gtool help


## Usage
The top level commands include:
* gorm
* help

## gtool version

    $ gtool -v

```
gtool version 0.0.1 darwin/amd64
```


## gtool gorm

    $ gtool gorm [-json] driverName datasourceName tableName [generatedPath]

 *Options:*

 * -json
    - true or false(default)
 * driverName
 * datasourceName
 * tableName
 * generatedPath

 **MYSQL corresponding type generation mapping description**

 |MYSQL|Type|GO|range|
 |:----    |:---|:----- |-----   |
 |int |signed  |int32 |-2147483648 ~ 2147483647   |
 |int |unsigned  |uint32 | 0 ~ 4294967295    |
 |smallint |signed  |int16 | -32768 ~ 32767    |
 |smallint |unsigned  |uint16 | 0 ~ 65535    |
 |tinyint |signed  |int8 | -128 ~ 127    |
 |tinyint |unsigned  |byte(uint8) | 0 ~ 255    |
 |bigint |signed  |int64 | -9223372036854776808 ~ 9223372036854775807    |
 |bigint |unsigned  |uint64 | 0 ~ 18446744073709551615    |
 |mediumint |signed  |int32 | -2147483648 ~ 2147483647    |
 |mediumint |unsigned  |uint32 | 0 ~ 4294967295    |
 |float |--  |float64 | --    |
 |double |--  |float64 | --    |
 |decimal |--  |string |--   |
 |date |--  |string |--    |
 |datetime |--  |time.Time |--    |
 |time |--  |string |--    |
 |varchar |--  |string |--    |
 |char |--  |string |--    |
 |text |--  |string |--    |

**Example**

    $ gtool gorm mysql "root:123456@tcp(127.0.0.1:3306)/test?charset=utf8" user ./model/

The generated `user.go` file is as follows：

```go
package model

type User struct {
    Id      int64   `gorm:"column:id;primary_key;AUTO_INCREMENT"`
    Name    string  `gorm:"column:name"`
    TestA   string  `gorm:"column:test_a"`
}
```