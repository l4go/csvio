type Reader
===
selectでの非同期処理に対応した、io.Readerからの行をカラムベースとして読み取り処理を行うライブラリです。  

## import

```go
import "github.com/l4go/csvio"
```
vendoringして使うことを推奨します。  

## 利用サンプル

[example](../examples/ex_csvio_r/ex_csvio_r.go)

## メソッド概略

### `func NewReader(rio io.Reader) (*Reader, error)`
Readerを生成します。  

### `func NewReaderWithConfig(rio io.Reader, conf *Config) (*Reader, error)`
Readerを、[csvio.Config](./Config.md) によるフォーマット指定により生成します。  

### `func (r *Reader) Recv() <-chan []string`
読み取ったデータをカラム毎に配列の要素として返すchannelを返します。  
読み取り完了もしくは、Readerの開放または、読み取りのエラーが発生した場合に、channelがクローズ状態になります。
エラーの発生は、クローズ状態のあと、`func (r *Reader) Err() error`メソッドを利用して取得します。

### `func (r *Reader) Err() error`
読み込みがcloseした場合(EOFやErrClosed)には、nilを返します。それ以外の場合は、該当のエラーの値を返します。
読み取りが正常終了を判定するために使います。

### `func (r *Reader) Close()`

Readerを開放するための後処理をします。
