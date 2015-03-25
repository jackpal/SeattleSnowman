// Copyright (C) 2015 John Howard Palevich. All Rights Reserved.

package db

import (
	"database/sql"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type sqldb struct {
	path string
	db   *sql.DB
}

func NewSQLDB(path string) (d DB) {
	return &sqldb{path, nil}
}

func (s *sqldb) Open() (err error) {
	db, err := sql.Open("sqlite3", s.path)
	if err != nil {
		return
	}
	s.db = db
	err = s.createTables()
	return
}

func (s *sqldb) createTables() (err error) {
	sqlStmt := `
	create table if not exists devices (ip text not null primary key, name text, activeUntil integer);
	`
	_, err = s.db.Exec(sqlStmt)
	return
}

func (s *sqldb) Add(d Device) (err error) {
	return s.AddAll([]Device{d})
}

func (s *sqldb) AddAll(devices []Device) (err error) {
	if len(devices) == 0 {
		return
	}

	tx, err := s.db.Begin()
	if err != nil {
		return
	}
	stmt, err := tx.Prepare("insert into devices(ip, name, activeUntil) values(?, ?, ?)")
	if err != nil {
		return
	}
	defer stmt.Close()
	for _, d := range devices {
		_, err = stmt.Exec(d.IP.String(), d.Name, timeToSQL(d.ActiveUntil))
		if err != nil {
			break
		}
	}
	tx.Commit()
	return
}

func (s *sqldb) Find(ip DeviceIP) (device Device, found bool, err error) {
	row := s.db.QueryRow("select name, activeUntil from devices where ip = ?",
		ip.String())
	var name string
	var activeUntil sql.NullInt64
	err = row.Scan(&name, &activeUntil)
	if err != nil {
		if err == sql.ErrNoRows {
			err = nil
		}
		return
	}
	found = true
	device = Device{ip, name, SQLToTime(activeUntil)}
	return
}

func (s *sqldb) Remove(ip DeviceIP) (err error) {
	_, err = s.db.Exec("delete from devices where ip is ?", ip.String())
	return
}

func (s *sqldb) All() (devices []Device, err error) {
	rows, err := s.db.Query("select ip, name, activeUntil from devices")
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var ip, name string
		var activeUntil sql.NullInt64
		rows.Scan(&ip, &name, &activeUntil)
		device := Device{ParseDeviceIP(ip), name, SQLToTime(activeUntil)}
		devices = append(devices, device)
	}
	rows.Close()
	return
}

func (s *sqldb) ModifyActiveUntil(ip DeviceIP, delta time.Duration, baseTime time.Time) (err error) {
	tx, err := s.db.Begin()
	if err != nil {
		return
	}
	var oldActiveTimeSQL sql.NullInt64
	err = tx.QueryRow("select activeUntil from devices where ip=?", ip.String()).Scan(&oldActiveTimeSQL)
	if err != nil {
		tx.Rollback()
		return
	}
	oldActiveTime := SQLToTime(oldActiveTimeSQL)
	newActiveTime := modifyActiveUntil(oldActiveTime, delta, baseTime)
	_, err = tx.Exec("update devices set activeUntil=? where ip=?", timeToSQL(newActiveTime), ip.String())
	if err != nil {
		tx.Rollback()
		return
	}
	err = tx.Commit()
	return
}

func (s *sqldb) SetActiveUntil(ip DeviceIP, activeUntil time.Time) (err error) {
	_, err = s.db.Exec("update devices set activeUntil=? where ip=?",
		timeToSQL(activeUntil), ip.String())
	return
}

func (s *sqldb) Close() (err error) {
	err = s.db.Close()
	return
}

func timeToSQL(t time.Time) sql.NullInt64 {
	if t.IsZero() {
		return sql.NullInt64{}
	}
	return sql.NullInt64{t.Unix(), true}
}

func SQLToTime(ni sql.NullInt64) time.Time {
	if !ni.Valid {
		return time.Time{}
	}
	return time.Unix(ni.Int64, 0)
}
