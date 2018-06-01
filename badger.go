package main

import (
	"log"
  "fmt"
	"github.com/dgraph-io/badger"
  "bytes"
  "encoding/gob"
)

type User struct {
  UserName string
  Password string
  Email string
}

func main() {
  // Open the Badger database located in the /tmp/badger directory.
  // It will be created if it doesn't exist.
  opts := badger.DefaultOptions
  opts.Dir = "/tmp/badger/azorg_logs/users"
  opts.ValueDir = "/tmp/badger/azorg_logs/users"
  db, err := badger.Open(opts)
  if err != nil {
	  log.Fatal(err)
  }
  defer db.Close()


  err1 := db.View(func(txn *badger.Txn) error {
    opts := badger.DefaultIteratorOptions
    opts.PrefetchValues = true
    it := txn.NewIterator(opts)
    for it.Rewind(); it.Valid(); it.Next() {
      k := it.Item().Key()
      v, err := it.Item().Value()
      if err != nil {
        return err
      }
      decBuf := bytes.NewBuffer(v)
      user := User{}
      err = gob.NewDecoder(decBuf).Decode(&user)
      // fmt.Printf("key=%s\nvalue=%s\n", k, v)
      fmt.Printf("Key: %s\nValue: %s\n", string(k), user.Password)
    }
    return nil
  })
  if err1 != nil {
    log.Fatal(err1)
  }
}

  // err1 := db.View(func(txn *badger.Txn) error {
  //   opts := badger.DefaultIteratorOptions
  //   opts.PrefetchValues = true
  //   it := txn.NewIterator(opts)
  //   for it.Rewind(); it.Valid(); it.Next() {
  //     k := it.Item().Key()
  //     v, err := it.Item().Value()
  //     if err != nil {
  //       return err
  //     }

  //     fmt.Printf("key=%s\nvalue=%s\n", k, v)
  //   }
  //   return nil
  // })
  // if err1 != nil {
  //   log.Fatal(err1)
  // }
