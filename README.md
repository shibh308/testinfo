# testinfo
パッケージ中にあるTest関数やExample関数を抽出し、テスト先の関数の情報や関数の呼び出し箇所を表示するツールです

# Install
```sh
$ go get github.com/shibh308/testinfo/cmd/testinfo
```

# Example
```sh
$ go vet -vettool=$(which testinfo) net/http
$ go vet -vettool=$(which testinfo) -testinfo.testfunc="TestNewClientServerTest" net/http
$ go vet -vettool=$(which testinfo) -testinfo.testfile="example_test.go" net/http
```
