Hello! My name is Huy Tran and I am a senior Math/CS major. This is my summer project.

Install go, and gosqlite

```
go get github.com/mattn/go-sqlite3
```

May have to do "go mod init" before this.

How to run the server locally:

```
go run *.go -c ../Examples/gem_config_local.json -add_teachers ../Examples/teachers.txt -add_students ../Examples/students.txt
```

To make sure the server run locally, open this URL on your browser:

```
http://127.0.0.1:8080/ping
```

You should see "pong".
