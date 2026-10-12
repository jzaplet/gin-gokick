package telemetry

import (
	"context"
	"regexp"

	"github.com/exaring/otelpgx"
)

var sqlcName = regexp.MustCompile(`^-- name: (\w+) :`)

func queryName(ctx context.Context, sql string) string {
	if match := sqlcName.FindStringSubmatch(sql); match != nil {
		return match[1]
	}

	return otelpgx.SQLOperationName(ctx, sql)
}
