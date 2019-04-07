package database

import (
    "log"
    "time"
)

type Workcenter struct {
    Workcenter string `db:"workcenter"json:"workcenter"`
    Description string `db:"description"json:"description"`
}

func (db *DB) Workcenters() ([]*Workcenter, error) {
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
    var ws []*Workcenter
    const query = "SELECT w.workcenter, w.description FROM workcenters w WHERE EXISTS (SELECT 1 FROM operations_clean ops WHERE ops.workcenter = w.workcenter LIMIT 1);"
    if err := tx.Select(&ws, tx.Rebind(query)); err != nil {
        log.Printf("Failed to select workcenters: %v", err)
        return nil, err
    }
    return ws, nil
}

type WorkcenterSequenceCount struct {
    WorkcenterStart *string `db:"workcenter_start"json:"workcenter_start"`
    WorkcenterEnd string `db:"workcenter_end"json:"workcenter_end"`
    Count int `db:"count"json:"count"`
}

func (db *DB) WorkcenterSequenceCount() ([]*WorkcenterSequenceCount, error) {
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
    var seqs []*WorkcenterSequenceCount
    const query = `SELECT
ops.prev_workcenter as workcenter_start, ops.workcenter as workcenter_end, count(*) as count
FROM operations_clean ops
WHERE EXISTS (SELECT 1 FROM workcenters WHERE workcenter = ops.workcenter LIMIT 1)
GROUP BY 1, 2;`
    if err := tx.Select(&seqs, tx.Rebind(query)); err != nil {
        log.Printf("Failed to select workcenters: %v", err)
        return nil, err
    }
    return seqs, nil
}

type WorkOvertime struct {
    TS time.Time `db:"ts" json:"ts"`
    Workcenter string `db:"workcenter" json:"workcenter"`
    WorkedTime time.Duration `db:"worked_time" json:"worked_time"`
    AllowedTime time.Duration `db:"allowed_time" json:"allowed_time"`
    AllowedTimeSeconds int64 `db:"allowed_time_seconds" json:"allowed_time_seconds"`
    OverTimeLimit time.Duration `db:"over_time_limit" json:"over_time_limit"`
    OverTimeLimitSeconds int64 `db:"over_time_limit_seconds" json:"over_time_limit_seconds"`
    OverTime time.Duration `db:"over_time" json:"over_time"`
    OverTimeSeconds time.Duration `db:"over_time_seconds" json:"over_time_seconds"`
    ExceedingOverTime bool `db:"exceeding_over_time" json:"exceeding_over_time"`
}

func (db *DB) WorkcenterOvertimeWeekly() ([]*WorkOvertime, error) {
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
    var ots []*WorkOvertime
    const query = `SELECT time_bucket_gapfill(interval '1 week', ops.start_time, '2018-01-01 00:00:00', '2018-12-31 23:59:59') as ts, ops.workcenter as workcenter,
       COALESCE(SUM(extract(epoch from ops.end_time - ops.start_time)), 0) as worked_time,
       (COALESCE(AVG(w.capacity / 100 * c.total), 0) * 60 * 60)::INTEGER as allowed_time_seconds,
       (COALESCE(AVG(c.week_overtime), 0) * 60 * 60)::INTEGER as over_time_limit_seconds,
       (COALESCE(SUM(extract(epoch from ops.end_time - ops.start_time)), 0) - COALESCE(AVG(w.capacity / 100 * c.total), 0) * 60 * 60)::INTEGER as over_time_seconds,
       COALESCE(SUM(extract(epoch from ops.end_time - ops.start_time)), 0) > 60 * 60 * COALESCE(AVG(w.capacity / 100 * c.total) + AVG(c.week_overtime), 0) as exceeding_over_time
FROM operations_clean ops
       JOIN workcenters w ON ops.workcenter = w.workcenter
       FULL JOIN calendar_models c ON w.calender = c.calender
WHERE ops.start_time > '2018-01-01 00:00:00' AND ops.start_time < '2018-12-31 23:59:59'
GROUP BY 1, 2, c.week_overtime
HAVING COALESCE(SUM(extract(epoch from ops.end_time - ops.start_time)), 0) <= 60 * 60 * COALESCE(AVG(w.capacity / 100 * c.total) + AVG(c.week_overtime), 0);`
    if err := tx.Select(&ots, tx.Rebind(query)); err != nil {
        log.Printf("Failed to select overtimes: %v", err)
        return nil, err
    }
    // Fix overtime seconds/DB innacuracy
    for _, ot := range ots {
        fixOvertimeDuration(ot)
    }
    return ots, nil
}

func fixOvertimeDuration(ot *WorkOvertime) {
    if ot.AllowedTime.Seconds() != 0 && ot.AllowedTimeSeconds != 0 {
        ot.AllowedTime = time.Duration(ot.AllowedTimeSeconds)
        ot.AllowedTimeSeconds = 0
    }
    if ot.OverTimeLimit.Seconds() != 0 && ot.OverTimeLimitSeconds != 0 {
        ot.OverTimeLimit = time.Duration(ot.OverTimeLimitSeconds)
        ot.OverTimeLimitSeconds = 0
    }
    if ot.OverTime.Seconds() != 0 && ot.OverTimeSeconds != 0 {
        ot.OverTime = time.Duration(ot.OverTimeSeconds)
        ot.OverTimeSeconds = 0
    }
}

func (db *DB) WorkcenterOvertimeDaily() ([]*WorkOvertime, error) {
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
    var ots []*WorkOvertime
    const query = `SELECT time_bucket_gapfill(interval '1 day', ops.start_time, '2018-01-01 00:00:00', '2018-12-31 23:59:59') as ts, ops.workcenter as workcenter,
COALESCE(SUM(extract(epoch from ops.end_time - ops.start_time)), 0) as worked_time,
(COALESCE(AVG(w.capacity / 100 * c.total), 0) * 60 * 60)::INTEGER as allowed_time_seconds,
(COALESCE(AVG(c.day_overtime), 0) * 60 * 60)::INTEGER as over_time_limit_seconds,
(COALESCE(SUM(extract(epoch from ops.end_time - ops.start_time)), 0) - COALESCE(AVG(w.capacity / 100 * c.total / 5) * 60 * 60, 0))::INTEGER as over_time_seconds,
COALESCE(SUM(extract(epoch from ops.end_time - ops.start_time)), 0) > COALESCE(AVG(w.capacity / 100 * c.total / 5) * 60 * 60 + AVG(c.day_overtime), 0) as exceeding_over_time
FROM operations_clean ops
JOIN workcenters w ON ops.workcenter = w.workcenter
FULL JOIN calendar_models c ON w.calender = c.calender
WHERE ops.start_time > '2018-01-01 00:00:00' AND ops.start_time < '2018-12-31 23:59:59'
GROUP BY 1, 2, c.day_overtime
HAVING COALESCE(SUM(extract(epoch from ops.end_time - ops.start_time)), 0) <= 60 * 60 * COALESCE(AVG(w.capacity / 100 * c.total) + AVG(c.day_overtime), 0);`
    if err := tx.Select(&ots, tx.Rebind(query)); err != nil {
        log.Printf("Failed to select overtimes: %v", err)
        return nil, err
    }
    // Fix overtime seconds/DB innacuracy
    for _, ot := range ots {
        fixOvertimeDuration(ot)
    }
    return ots, nil
}



