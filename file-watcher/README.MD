# File Watcher

Lib to execute commands when file changes.

## Usage

```shell
file-watcher --files "example/*.sh" --delay 3 --signal SIGINT --command "echo ""file changed!""" --command "./example/test.sh"
```

Is important to note that Go escapes `"` using another `"` and not the usual backslash `\`.

## Dependencies

- [fsnotify](https://github.com/fsnotify/fsnotify)
