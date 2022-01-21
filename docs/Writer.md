type Writer
===
selectでの非同期処理に対応した、io.WriteCloserへの行をカラムベースな書き込み処理を行うライブラリです。  

## import

```go
import "github.com/l4go/csvio"
```
vendoringして使うことを推奨します。  

## 利用サンプル

[example](../examples/ex_csvio_w/ex_csvio_w.go)

## メソッド概略

### `func NewWriter(wio io.WriteCloser) (*Writer, error)`
Writerを生成します。  

### `func NewWriterWithConfig(wio io.WriteCloser, cnf *Config) (*Writer, error)`
Writerを、[csvio.Config](./Config.md) によるフォーマット指定により生成します。  

### `func (w *Writer) Send() chan<- []string`
カラム毎に配列の要素として渡すと書き込みを行うchannelを返します。  
Writerの開放処理(`func (w *Writer) Close()`もしくは、`func (w *Writer) Cancel()`) または、書き込みのエラーが発生した場合に、channelがクローズ状態になります。  
エラーの発生は、クローズ状態のあと、`func (w *Writer) Err() error`メソッドを利用して取得します。

### `func (w *Writer) Err() error`
読み込みがcloseした場合(EOFやErrClosed)には、nilを返します。それ以外の場合は、該当のエラーの値を返します。
読み取りが正常終了を判定するために使います。

### `func (w *Writer) Cancel()`

現在の書き込みキューに含まれるデータを書き込み切らず、Writerを開放します。  
行が中途半端に終了する可能性があります。  

### `func (w *Writer) Close()`

現在の書き込みキューに含まれるデータを書き込んでから、Writerを開放します。
