# Lightweight logger for Go

This library is a simple, lightweight logger for Go. By design, it is limited in
scope in order to make it simple to use, requiring virtually no configuration.

**Licence**: GPLv3

A trivial example:

```go
lg := lwlog.NewStd()
lg.Infof("an integer: %d", 123)
```

The interface `lwlog.Logger` is intended to be used by any other package that
requires logging functionality.

The “standard” logger (`lwlog.Std`) included in this package will log
debug/info to `stdout` and errors to `stderr`. There is also the `lwjournal`
package (in its [own repository](https://github.com/lwithers/lwjournal), to
avoid pulling unnecessary dependencies into this package) which provides a
`lwjournal.Logger` object that implements the `lwlog.Logger` interface; this
specific implementation logs to systemd's
[journal](https://www.freedesktop.org/software/systemd/man/systemd-journald.service.html).
