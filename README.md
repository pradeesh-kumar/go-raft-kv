# Go-Raft-KV

The project go-raft-kv is a Distributed In-Memory Key-Value store implemented using Raft consensus algorithm.

# Features

* Leader Election
* Log Replication
* Leadership Transfer
* Snapshot
* Dynamic Membership & Configuration Change
* Raft Client

# Build Protobuf

`
cd server
protoc raft/*.proto --go_out=. --go-grpc_out=. --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative --proto_path=.
protoc kv/*.proto --go_out=. --go-grpc_out=. --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative --proto_path=.
`

# Build App

go build server_main.go

# Run

`
go run server_main.go -serverId node_1 -serverAddress localhost:7071
go run server_main.go -serverId node_2 -serverAddress localhost:7072
go run server_main.go -serverId node_3 -serverAddress localhost:7073
`

To run with verbose mode use the flag -v

# References

* Raft Paper: https://raft.github.io/raft.pdf
* Raft Web Page: https://raft.github.io/
* Raft dissertation: https://github.com/ongardie/dissertation/blob/master/book.pdf?raw=true