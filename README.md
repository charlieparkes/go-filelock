# go-filelock

```golang
lockPath := filepath.Join(homeDir, "test_pgsql.lock")


myLock, err := filelock.Lock(lockPath)
if err == nil {
    defer errors.Ignore(myLock.Unlock)
    // do thing
    if err := myLock.Unlock(); err != nil {
        panic(err)
    }
}

// ALSO

if exists, _ := filelock.Exists(lockPath); exists {
    // do thing
}
```