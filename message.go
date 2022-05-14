package sqldblogger

const (
	MessageCommit              = "Commit"
	MessageRollback            = "Rollback"
	MessageResultLastInsertId  = "ResultLastInsertId"
	MessageResultRowsAffected  = "ResultRowsAffected"
	MessageBegin               = "Begin"
	MessagePrepareContext      = "PrepareContext"
	MessagePrepare             = "Prepare"
	MessageClose               = "Close"
	MessageConnect             = "Connect"
	MessageBeginTx             = "BeginTx"
	MessagePing                = "Ping"
	MessageExec                = "Exec"
	MessageExecContext         = "ExecContext"
	MessageQuery               = "Query"
	MessageQueryContext        = "QueryContext"
	MessageResetSession        = "ResetSession"
	MessageCheckNamedValue     = "CheckNamedValue"
	MessageRowsClose           = "RowsClose"
	MessageRowsNext            = "RowsNext"
	MessageRowsNextResultSet   = "RowsNextResultSet"
	MessageStmtClose           = "StmtClose"
	MessageStmtExec            = "StmtExec"
	MessageStmtQuery           = "StmtQuery"
	MessageStmtCheckNamedValue = "StmtCheckNamedValue"
	MessageStmtQueryContext    = "StmtQueryContext"
	MessageStmtExecContext     = "StmtExecContext"
)

var mapMsgToLevel = map[string]Level{
	MessageCommit:              LevelDebug,
	MessageRollback:            LevelDebug,
	MessageResultLastInsertId:  LevelTrace,
	MessageResultRowsAffected:  LevelTrace,
	MessageBegin:               LevelDebug,
	MessagePrepare:             LevelInfo,
	MessagePrepareContext:      LevelInfo,
	MessageClose:               LevelDebug,
	MessageBeginTx:             LevelDebug,
	MessagePing:                LevelDebug,
	MessageExec:                LevelDebug,
	MessageExecContext:         LevelDebug,
	MessageQuery:               LevelInfo,
	MessageQueryContext:        LevelDebug,
	MessageResetSession:        LevelTrace,
	MessageCheckNamedValue:     LevelTrace,
	MessageRowsClose:           LevelTrace,
	MessageRowsNext:            LevelTrace,
	MessageRowsNextResultSet:   LevelTrace,
	MessageStmtClose:           LevelDebug,
	MessageStmtExec:            LevelDebug,
	MessageStmtExecContext:     LevelDebug,
	MessageStmtQuery:           LevelInfo,
	MessageStmtQueryContext:    LevelInfo,
	MessageStmtCheckNamedValue: LevelTrace,
}

func getDefaultLevelByMessage(msg string, level *Level) Level {
	if level != nil {
		return *level
	} else if lvl, ok := mapMsgToLevel[msg]; ok {
		return lvl
	}
	return LevelTrace
}
