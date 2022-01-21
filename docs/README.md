golib/csvio
===

selectでの非同期処理に対応した、csvベースでの読み込み/書き込みを行うライブラリです。  

* [csvio.Config](./Config.md)
	* CSVのコンマの値や、クォートの有無など、扱うCSVに対するフォーマットを指定できます。  
* [csvio.Reader](./Reader.md)
	* selectでの非同期処理に対応した、io.Readerからの行をカラムベースとして読み取り処理を行います。  
* [csvio.Writer](./Writer.md)
	* selectでの非同期処理に対応した、io.WriteCloserへの行をカラムベースな書き込み処理を行います。  
