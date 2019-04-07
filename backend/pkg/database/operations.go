package database

import (
    "log"
    "time"
)

type IdleOrder struct {
    TS time.Time `db:ts",json:"ts"`
    Workcenter string `db:"workcenter",json:"workcenter"`
    WaitingTime time.Duration `db:"waiting_time",json:"waiting_time"`
    PreparationTime time.Duration `db:"preparation_time",json:"preparation_time"`
    IdleTime time.Duration `db:"idle_time",json:"idle_time"`
}

func (db *DB) IdleOrders() ([]*IdleOrder, error) {
    tx, err := db.db.Beginx()
    if err != nil {
        log.Printf("Failed to begin transaction: %v", err)
        return nil, err
    }
    defer func() {
        if err := tx.Rollback(); err != nil {
            log.Printf("Failed to roll back transaction: %v", err)
        }
    }()
    var io []*IdleOrder
    const query = `SELECT time_bucket_gapfill(interval '31 days', ops.start_time, '2018-01-01 00:00:00', '2018-12-31 23:59:59') as ts, ops.workcenter as workcenter,
COALESCE(extract(epoch from SUM(ops_next.start_time - ops.end_time)), 0) as waiting_time,
COALESCE(SUM(ops.plan_setup_time_seconds + o.amount * ops.plan_processing_time_seconds), 0) as preparation_time,
COALESCE(extract(epoch from SUM(ops_next.start_time - ops.end_time)), 0) - COALESCE(SUM(ops.plan_setup_time_seconds + o.amount * ops.plan_processing_time_seconds), 0) as idle_time
FROM operations_clean ops
JOIN operations_clean ops_next
ON ops.order_id = ops_next.order_id
AND ops.workcenter = ops_next.workcenter
AND ops_next.start_time > ops.start_time
JOIN orders o ON ops.order_id = o.order_id
GROUP BY 1, 2
ORDER BY idle_time DESC;`
    if err := tx.Select(&io, tx.Rebind(query)); err != nil {
        log.Printf("Failed to select idle orders: %v", err)
        return nil, err
    }
    return io, nil
}


type SpecialOrder struct {
    TS time.Time `db:"ts",json:"ts"`
    Workcenter string `db:"workcenter",json:"workcenter"`
    Count int `db:"count",json:"count"`
}

func (db *DB) SpecialOrders() ([]*SpecialOrder, error) {
    tx, err := db.db.Beginx()
    if err != nil {
        log.Printf("Failed to begin transaction: %v", err)
        return nil, err
    }
    defer func() {
        if err := tx.Rollback(); err != nil {
            log.Printf("Failed to roll back transaction: %v", err)
        }
    }()
    var so []*SpecialOrder
    const query = `SELECT time_bucket_gapfill(interval '31 days', ops.start_time, '2018-01-01 00:00:00', '2018-12-31 23:59:59') as ts,
    w.workcenter as workcenter, COALESCE(COUNT(*), 0) as count 
FROM operations_clean ops
JOIN workcenters w ON ops.workcenter = w.workcenter
WHERE ops.order_id LIKE 'A%'
GROUP BY 1, 2;`
    if err := tx.Select(&so, tx.Rebind(query)); err != nil {
        log.Printf("Failed to select special orders: %v", err)
        return nil, err
    }
    return so, nil
}
