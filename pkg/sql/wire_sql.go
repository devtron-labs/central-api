package sql

import "github.com/google/wire"

var PgSqlWireSet = wire.NewSet(
	ParseConfig,
	NewDbConnection,
)
