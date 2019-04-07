package database

import (
    "log"
    "time"
)

type DurationLowerBound struct {
    StartTime time.Time `db:"start_time" json:"start_time"`
    OperationTime time.Duration `db:"operation_time" json:"operation_time"`
    OperationTimeSeconds int64 `db:"operation_time_seconds" json:"operation_time_seconds"`
    TransportTime time.Duration `db:"transport_time" json:"transport_time"`
    TransportTimeSeconds int64 `db:"transport_time_seconds" json:"transport_time_seconds"`
    TotalTime time.Duration `db:"total_time" json:"total_time"`
    TotalTimeSeconds int64 `db:"total_time_seconds" json:"total_time_seconds"`
    OrderTime time.Duration  `db:"order_time" json:"order_time"`
    OrderTimeSeconds int64 `db:"order_time_seconds" json:"order_time_seconds"`
}

func (db *DB) DurationLowerBound(orderID int64) ([]*DurationLowerBound, error) {
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
    var dlb []*DurationLowerBound
    var query string
    var args []interface{}
    if orderID != 0 {
        query = `SELECT ops.start_time as start_time, (ops.plan_setup_time_seconds + o.amount * ops.plan_processing_time_seconds)::INTEGER as operation_time_seconds,
  (CASE
    WHEN m.weight_gross >= 25 THEN 24 * 60 * 60
    ELSE 12 * 60 * 60
  END)::INTEGER as transport_time_seconds,
  (ops.plan_setup_time_seconds + o.amount * ops.processing_time_seconds +
  CASE
    WHEN m.weight_gross >= 25 THEN 24 * 60 * 60
    ELSE 12 * 60 * 60
  END)::INTEGER as total_time_seconds,
  ((extract(epoch from o.end_time) - extract(epoch from o.start_time)) / 60 / 60)::INTEGER as order_time_seconds
FROM operations_clean ops
JOIN orders o ON ops.order_id = o.order_id
JOIN materials m ON o.material = m.material
WHERE ops.order_id = ? AND ops.start_time > '2018-01-01 00:00:00' AND ops.start_time < '2018-12-31 23:59:59';`
        args = append(args, orderID)
    } else {
        query = `SELECT ops.start_time as start_time, ops.plan_setup_time_seconds + o.amount * ops.processing_time_seconds as operation_time_seconds,
  (CASE
    WHEN m.weight_gross >= 25 THEN (24 * 60 * 60)::INTEGER
    ELSE (12 * 60 * 60)::INTEGER
  END)::INTEGER as transport_time,
  (ops.plan_setup_time_seconds + o.amount * ops.processing_time_seconds +
  CASE
    WHEN m.weight_gross >= 25 THEN (24 * 60 * 60)::INTEGER
    ELSE (12 * 60 * 60)::INTEGER
  END)::INTEGER as total_time_seconds,
  ((extract(epoch from o.end_time) - extract(epoch from o.start_time)) / 60 / 60)::INTEGER as order_time_seconds
FROM operations_clean ops
JOIN orders o ON ops.order_id = o.order_id
JOIN materials m ON o.material = m.material
WHERE ops.start_time > '2018-01-01 00:00:00' AND ops.start_time < '2018-12-31 23:59:59';`
    }

    if err := tx.Select(&dlb, tx.Rebind(query), args...); err != nil {
        log.Printf("Failed to select duration lower bound: %v", err)
        return nil, err
    }
    // Fix overtime seconds/DB innacuracy
    for _, lb := range dlb {
        fixLowerBoundDuration(lb)
    }
    return dlb, nil
}

func fixLowerBoundDuration(lb *DurationLowerBound) {
    if lb.OperationTime.Seconds() != 0 && lb.OperationTimeSeconds != 0 {
        lb.OperationTime = time.Duration(lb.OperationTimeSeconds)
        lb.OperationTimeSeconds = 0
    }
    if lb.TransportTime.Seconds() != 0 && lb.TransportTimeSeconds != 0 {
        lb.TransportTime = time.Duration(lb.TransportTimeSeconds)
        lb.TransportTimeSeconds = 0
    }
    if lb.TotalTime.Seconds() != 0 && lb.TotalTimeSeconds != 0 {
        lb.TotalTime = time.Duration(lb.TotalTimeSeconds)
        lb.TotalTimeSeconds = 0
    }
    if lb.OrderTime.Seconds() != 0 && lb.OrderTimeSeconds != 0 {
        lb.OrderTime = time.Duration(lb.OrderTimeSeconds)
        lb.OrderTimeSeconds = 0
    }
}

type MissedOrderDeadline struct {
    OrderID string `db:"order_id" json:"order_id"`
    OrderStartTime time.Time `db:"order_start_time" json:"order_start_time"`
    OrderEndTime time.Time `db:"order_end_time" json:"order_end_time"`
    OperationsStartTime time.Time `db:"operations_start_time" json:"operations_start_time"`
    OperationsEndTime time.Time `db:"operations_end_time" json:"operations_end_time"`
    Late bool `db:"late" json:"late"`
}

func (db *DB) MissedOrderDeadlines() ([]*MissedOrderDeadline, error) {
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
    var mod []*MissedOrderDeadline
    const query = `SELECT ops.order_id as order_id, o.start_time as order_start_time, o.end_time as order_end_time,
ops.start_time as operations_start_time, ops.end_time as operations_end_time,
ops.end_time > o.end_time as late
FROM operations_clean ops
JOIN orders o ON ops.order_id = o.order_id
WHERE ops.end_time > o.end_time;`

    if err := tx.Select(&mod, tx.Rebind(query)); err != nil {
        log.Printf("Failed to select missed order deadlines: %v", err)
        return nil, err
    }
    return mod, nil
}

type OrderHistoryEntry struct {
    OrderID string `db:"order_id" json:"order_id"`
    Operation string `db:"operation" json:"operation"`
    PreviousOperation *int `db:"prev_operation" json:"prev_operation"`
    Workcenter string `db:"workcenter" json:"workcenter"`
    PrevWorkcenter *string `db:"prev_workcenter" json:"prev_workcenter"`
    StartTime time.Time `db:"start_time" json:"start_time"`
    EndTime time.Time `db:"end_time" json:"end_time"`
    SetupTime time.Duration `db:"setup_time" json:"setup_time"`
    ProcessingTime time.Duration `db:"processing_time" json:"processing_time"`
    PlanSetupTime time.Duration `db:"plan_setup_time" json:"plan_setup_time"`
    PlanProcessingTime time.Duration `db:"plan_processing_time" json:"plan_processing_time"`
}

func (db *DB) OrderHistory(orderID string) ([]*OrderHistoryEntry, error) {
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
    var ohe []*OrderHistoryEntry
    const query = `SELECT ops.order_id as order_id, ops.operation as operation, ops.prev_operation as prev_operation,
ops.workcenter as workcenter, ops.prev_workcenter as prev_workcenter, ops.start_time as start_time, ops.end_time as end_time,
extract(epoch from ops.setup_time)::INTEGER as setup_time, extract(epoch from ops.processing_time)::INTEGER as processing_time,
extract(epoch from ops.plan_setup_time)::INTEGER as plan_setup_time, extract(epoch from ops.plan_processing_time)::INTEGER as plan_processing_time
FROM operations_clean ops WHERE ops.order_id = ?
ORDER BY start_time, operation;`

    if err := tx.Select(&ohe, tx.Rebind(query), orderID); err != nil {
        log.Printf("Failed to select missed order deadlines: %v", err)
        return nil, err
    }
    return ohe, nil
}
