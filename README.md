# TipTop
A lightweight two-level cache (in-memory + redis) library for Go. FIFO is support when 
the capacity of cache in shard is exceed the setting, and the entry will be removed 
and stored in the redis which is optional on.

This library is using the lazy delete and the periodically delete strategy as the redis 
provided by goroutine. 

## How to use
```go
package you_package

import (
    "fmt"
	"github.com/guriytan/tiptop"
)

func main()  {
    t, _ := tiptop.NewTipTop(tiptop.DefaultConfig())
    value := []byte("test-value")
    err := t.Set("key1", value)
    if err != nil {
        fmt.Println(err.Error())
    }
    bytes, err := t.Get("key1")
    if err != nil {
        fmt.Println(err.Error())
        return
    }
    fmt.Printf("输出：%v\n", string(bytes))
}
```
or you can use custom config by this.
```go
t, err := tiptop.NewTipTop(tiptop.Config{
    ShardSize:    1024,
    MaxCacheSize: 50 * t.MB,
    OnRemove:    true,
})
```

## Performance
```shell script
goos: windows
goarch: amd64
pkg: guriytan.cn/tiptop
BenchmarkTipTop_Get
BenchmarkTipTop_Get/512-shards
BenchmarkTipTop_Get/512-shards-8         	15004651	        71.9 ns/op	      23 B/op	       2 allocs/op
BenchmarkTipTop_Get/1024-shards
BenchmarkTipTop_Get/1024-shards-8        	14632450	        74.8 ns/op	      23 B/op	       2 allocs/op
BenchmarkTipTop_Get/4096-shards
BenchmarkTipTop_Get/4096-shards-8        	11430432	       115 ns/op	      23 B/op	       2 allocs/op
PASS
BenchmarkTipTop_Set
BenchmarkTipTop_Set/512-shards
BenchmarkTipTop_Set/512-shards-8         	 7057590	       166 ns/op	      80 B/op	       4 allocs/op
BenchmarkTipTop_Set/1024-shards
BenchmarkTipTop_Set/1024-shards-8        	 7228971	       163 ns/op	      80 B/op	       4 allocs/op
BenchmarkTipTop_Set/4096-shards
BenchmarkTipTop_Set/4096-shards-8        	 5194530	       211 ns/op	      81 B/op	       4 allocs/op
PASS
```

## How it works

TipTop shards the cache container and entrusts read-write locks to improve the concurrency 
performance, so it is necessary to calculate the hash of key to select the corresponding shard 
container for storage and query.The value need to be serialized to []byte for storage.

## Reference
1. [cache](https://github.com/seaguest/cache)
2. [bigcache](https://github.com/allegro/bigcache)
2. [freecache](https://github.com/coocood/freecache)

## License
The Apache 2.0 license