# testinfo

```sh
$ go vet -vettool=`which testinfo` net/http
$ go vet -vettool=`which testinfo` -testinfo.testfunc="TestNewClientServerTest" net/http
$ go vet -vettool=`which testinfo` -testinfo.testfile="example_test.go" net/http
```
