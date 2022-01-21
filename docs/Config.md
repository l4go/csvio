type Config
===

CSVのコンマの値や、クォートの有無など、扱うCSVに対するフォーマットを指定できます。  
デフォルト値は、csvio.DefaultConfigに定義されている通りです。  

## import

```go
import "github.com/l4go/csvio"
```
vendoringして使うことを推奨します。  

## メソッド概略

### `func (cnf *Config) Check() error`
Config自体の設定にエラーがないか確認します。  
エラーがある場合、csvio.ErrConfigを返します。  
