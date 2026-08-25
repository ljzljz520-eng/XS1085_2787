package memorial

import "context"

type Repository interface {
	SaveMemorial(context.Context, Memorial) error
	GetMemorial(context.Context, string) (Memorial, error)
	SaveCandle(context.Context, Candle) error
	GetCandle(context.Context, string) (Candle, error)
	SaveSpark(context.Context, Spark) error
	ListSparks(context.Context, string, int) ([]Spark, error)
	SaveSession(context.Context, VisitorSession) error
	GetSession(context.Context, string) (VisitorSession, error)
	SaveReflection(context.Context, Reflection) error
	ListReflections(context.Context, string) ([]Reflection, error)
	SaveAudit(context.Context, AuditEntry) error
	CountVisitors(context.Context, string) (int, error)
}
