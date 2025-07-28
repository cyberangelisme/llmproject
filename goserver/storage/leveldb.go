// Package storage 提供 LevelDB 的简单封装
package storage

import (
	"fmt"
	"log"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
)

// LevelDBWrapper 是对 github.com/syndtr/goleveldb/leveldb.DB 的封装
type LevelDBWrapper struct {
	db *leveldb.DB
}

// NewLevelDBWrapper 创建一个新的 LevelDBWrapper 实例
// path: LevelDB 数据库文件的路径
// readOnly: 是否以只读模式打开
func NewLevelDBWrapper(path string, readOnly bool) (*LevelDBWrapper, error) {
	var db *leveldb.DB
	var err error

	if readOnly {
		// 以只读模式打开
		db, err = leveldb.OpenFile(path, &opt.Options{
			ReadOnly: true,
		})
	} else {
		// 以读写模式打开（如果数据库不存在会自动创建）
		db, err = leveldb.OpenFile(path, nil) // nil 表示使用默认选项
		// 或者显式指定选项:
		// db, err = leveldb.OpenFile(path, &opt.Options{
		//     // 可以在这里设置更多选项，例如:
		//     // BlockCacheCapacity: 100 * opt.MiB, // 设置块缓存大小
		//     // WriteBuffer: 64 * opt.MiB,        // 设置写入缓冲区大小
		// })
	}

	if err != nil {
		return nil, fmt.Errorf("failed to open LevelDB at %s: %w", path, err)
	}

	return &LevelDBWrapper{
		db: db,
	}, nil
}

// Put 将键值对存入数据库
func (l *LevelDBWrapper) Put(key, value []byte) error {
	// 使用默认的写入选项
	err := l.db.Put(key, value, nil)
	if err != nil {
		return fmt.Errorf("failed to put key %s: %w", string(key), err)
	}
	return nil
}

// Get 从数据库中获取值
// 如果键不存在，会返回 leveldb.ErrNotFound 错误
func (l *LevelDBWrapper) Get(key []byte) ([]byte, error) {
	// 使用默认的读取选项
	value, err := l.db.Get(key, nil)
	if err != nil {
		// 检查是否是未找到的错误
		if err == leveldb.ErrNotFound {
			// 可以选择返回 nil, nil 或者原样返回错误
			// 这里选择返回错误，让调用者决定如何处理
			return nil, fmt.Errorf("key %s not found: %w", string(key), err)
		}
		return nil, fmt.Errorf("failed to get key %s: %w", string(key), err)
	}
	// 注意：返回的 value 是 leveldb 内部 buffer 的拷贝，可以直接使用
	// 如果需要长时间持有，建议进行深拷贝: result := append([]byte(nil), value...)
	return value, nil
}

// Has 检查键是否存在
func (l *LevelDBWrapper) Has(key []byte) (bool, error) {
	exists, err := l.db.Has(key, nil)
	if err != nil {
		return false, fmt.Errorf("failed to check key %s existence: %w", string(key), err)
	}
	return exists, nil
}

// Delete 从数据库中删除键值对
func (l *LevelDBWrapper) Delete(key []byte) error {
	err := l.db.Delete(key, nil)
	if err != nil {
		return fmt.Errorf("failed to delete key %s: %w", string(key), err)
	}
	return nil
}

// Close 关闭数据库连接
// 应用结束前务必调用此方法
func (l *LevelDBWrapper) Close() error {
	if l.db != nil {
		err := l.db.Close()
		if err != nil {
			return fmt.Errorf("failed to close LevelDB: %w", err)
		}
		log.Println("LevelDB closed successfully.")
	}
	return nil
}

// NewIterator 创建一个新的迭代器
// slice: 可选的 key 范围限制，nil 表示遍历所有 key
// 用完迭代器后，务必调用 iter.Release()
// 示例:
// iter := db.NewIterator(nil)
// defer iter.Release()
//
//	for iter.Next() {
//	    key := iter.Key()
//	    value := iter.Value()
//	    // ... 处理 key/value ...
//	}
//
//	if err := iter.Error(); err != nil {
//	    // ... 处理错误 ...
//	}
func (l *LevelDBWrapper) NewIterator(slice *util.Range) *leveldb.Iterator {
	return l.db.NewIterator(slice, nil)
}

// CompactRange 对指定范围的 key 进行压缩，释放磁盘空间
// 使用 util.Range{} 表示压缩整个数据库
func (l *LevelDBWrapper) CompactRange(r util.Range) error {
	err := l.db.CompactRange(r)
	if err != nil {
		return fmt.Errorf("failed to compact range: %w", err)
	}
	return nil
}

// 示例：如何使用这个封装
/*
func main() {
    // 1. 打开或创建数据库
    db, err := NewLevelDBWrapper("./data/my_leveldb", false)
    if err != nil {
        log.Fatalf("Failed to open database: %v", err)
    }
    defer db.Close() // 确保程序退出时关闭数据库

    // 2. 存储数据
    err = db.Put([]byte("key1"), []byte("value1"))
    if err != nil {
        log.Printf("Put error: %v", err)
    } else {
        log.Println("Put key1=value1")
    }

    err = db.Put([]byte("key2"), []byte("value2"))
    if err != nil {
        log.Printf("Put error: %v", err)
    } else {
        log.Println("Put key2=value2")
    }

    // 3. 获取数据
    value, err := db.Get([]byte("key1"))
    if err != nil {
        if err.Error() == "key key1 not found: leveldb: not found" {
            log.Println("Key1 not found")
        } else {
            log.Printf("Get error: %v", err)
        }
    } else {
        log.Printf("Get key1: %s", string(value))
    }

    // 4. 检查键是否存在
    exists, err := db.Has([]byte("key2"))
    if err != nil {
        log.Printf("Has error: %v", err)
    } else {
        log.Printf("Key2 exists: %t", exists)
    }

    // 5. 删除数据
    err = db.Delete([]byte("key1"))
    if err != nil {
        log.Printf("Delete error: %v", err)
    } else {
        log.Println("Deleted key1")
    }

    // 6. 再次获取已删除的数据
    _, err = db.Get([]byte("key1"))
    if err != nil {
        log.Printf("Get deleted key1 error (expected): %v", err)
    }

    // 7. 遍历所有数据
    log.Println("Iterating all keys:")
    iter := db.NewIterator(nil)
    defer iter.Release()
    for iter.Next() {
        key := iter.Key()
        value := iter.Value()
        log.Printf("  %s -> %s", string(key), string(value))
    }
    if err := iter.Error(); err != nil {
        log.Printf("Iterator error: %v", err)
    }
}
*/
