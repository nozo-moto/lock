package lock

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type MySQLLocker struct {
	key  string
	conn *sql.Conn
}

func NewMySQLLocker(conn *sql.Conn) *MySQLLocker {
	return &MySQLLocker{
		key:  lockKey,
		conn: conn,
	}
}

func (m *MySQLLocker) GetLock(ctx context.Context, timeout time.Duration) error {
	var res int64
	err := m.conn.QueryRowContext(
		ctx,
		"SELECT GET_LOCK(?, ?)",
		m.key,
		int(timeout.Seconds()),
	).Scan(&res)
	if err != nil {
		return err
	}

	if res != 1 {
		return fmt.Errorf("timeout to get lock %q", m.key)
	}

	return nil
}

func (m *MySQLLocker) HasLock(ctx context.Context) (bool, error) {
	row := m.conn.QueryRowContext(ctx, "SELECT IS_USED_LOCK(?) = CONNECTION_ID()", m.key)

	var res sql.NullInt64
	err := row.Scan(&res)
	if err != nil {
		return false, err
	}

	if !res.Valid {
		return false, nil
	}

	return res.Int64 == 1, nil
}

func (m *MySQLLocker) ReleaseLock(ctx context.Context) error {
	var res sql.NullInt64
	err := m.conn.QueryRowContext(ctx, "SELECT RELEASE_LOCK(?)", m.key).Scan(&res)
	if err != nil {
		return err
	}

	if !res.Valid || res.Int64 != 1 {
		return fmt.Errorf("failed to release %q", m.key)
	}

	return err
}
