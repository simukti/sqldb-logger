module github.com/simukti/sqldb-logger/logadapter/onelogadapter

go 1.15

require (
	github.com/francoispqt/onelog v0.0.0-20190306043706-8c2bb31b10a4
	github.com/simukti/sqldb-logger v0.0.0-20200812042017-c462204a3317
	github.com/stretchr/testify v1.6.1
)

replace github.com/simukti/sqldb-logger => ../../
