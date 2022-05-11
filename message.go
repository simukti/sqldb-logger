package sqldblogger

const initLogMsg = "{}"

const (
	MessageCommit              = "Commit"
	MessageRollback            = "Rollback"
	MessageResultLastInsertId  = "ResultLastInsertId"
	MessageResultRowsAffected  = "ResultRowsAffected"
	MessageBegin               = "Begin"
	MessagePrepare             = "Prepare"
	MessageClose               = "Close"
	MessageBeginTx             = "BeginTx"
	MessagePrepareContext      = "PrepareContext"
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
	MessagePrepare:             LevelDebug,
	MessageClose:               LevelDebug,
	MessageBeginTx:             LevelDebug,
	MessagePrepareContext:      LevelDebug,
	MessagePing:                LevelDebug,
	MessageExec:                LevelInfo,
	MessageExecContext:         LevelDebug,
	MessageQuery:               LevelDebug,
	MessageQueryContext:        LevelDebug,
	MessageResetSession:        LevelTrace,
	MessageCheckNamedValue:     LevelTrace,
	MessageRowsClose:           LevelTrace,
	MessageRowsNext:            LevelTrace,
	MessageRowsNextResultSet:   LevelTrace,
	MessageStmtClose:           LevelDebug,
	MessageStmtExec:            LevelDebug,
	MessageStmtExecContext:     LevelDebug,
	MessageStmtQuery:           LevelDebug,
	MessageStmtQueryContext:    LevelDebug,
	MessageStmtCheckNamedValue: LevelTrace,
}

func getDefaultLevelByMessage(msg string) Level {
	if lvl, ok := mapMsgToLevel[msg]; ok {
		return lvl
	}
	return LevelTrace
}

func isAbleToPrinted(o *options, msg string, lvl Level) bool {
	var myLevel Level
	switch msg {
	case MessagePrepare, MessagePrepareContext:
		myLevel = o.preparerLevel
		break
	case MessageExecContext, MessageExec, MessageStmtExec, MessageStmtExecContext:
		myLevel = o.execerLevel
		break
	case MessageQuery, MessageQueryContext, MessageStmtQuery, MessageStmtQueryContext:
		myLevel = o.queryerLevel
		break
	default:
		myLevel = o.minimumLogLevel
		break
	}
	return myLevel <= lvl
}
