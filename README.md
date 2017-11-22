remotepmyamux
=============

Reverse portmapper

Build
=====

go get -u github.com/eelf/remotepmyamux

env GOARCH=amd64 GOOS=linux sh -c 'go install -pkgdir=$GOPATH/pkg/$GOOS-$GOARCH github.com/eelf/remotepmyamux'

go install github.com/eelf/remotepmyamux

Usage
=====

Wait for remote host at :3307 and expose local port 3306 which will be connected with service at remote host
remotepmyamux -control-addr=:3307 -local-addr=:3306

Connect to remote host at example.com:3307 and pass its local connects to this host unix socket

remotepmyamux -control-addr=example.com:3307 -service-addr=/var/run/mysql.sock -service-net=unix
