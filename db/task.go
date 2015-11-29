/**
 * @file task.go
 * @author Mikhail Klementyev jollheef<AT>riseup.net
 * @license GNU GPLv3
 * @date November, 2015
 * @brief queries for task table
 */

package db

import (
	"database/sql"
)

// Task row
type Task struct {
	ID            int
	Name          string
	Desc          string
	CategoryID    int
	Level         int
	Price         int
	Shared        bool
	Flag          string
	MaxSharePrice int
	MinSharePrice int
	Opened        bool
}

func createTaskTable(db *sql.DB) (err error) {

	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS "task" (
		id		SERIAL PRIMARY KEY,
		name		TEXT NOT NULL,
		description	TEXT NOT NULL,
		category_id	INTEGER NOT NULL,
		level		INTEGER NOT NULL,
		price		INTEGER NOT NULL,
		shared		BOOLEAN NOT NULL,
		flag		TEXT NOT NULL,
		max_share_price	INTEGER NOT NULL,
		min_share_price	INTEGER NOT NULL,
		opened		BOOLEAN NOT NULL
	)`)

	return
}

// AddTask add task and fill id
func AddTask(db *sql.DB, t *Task) (err error) {

	stmt, err := db.Prepare("INSERT INTO task (name, description, " +
		"category_id, level, price, shared, flag, max_share_price, " +
		"min_share_price, opened) VALUES ($1, $2, $3, $4, $5, $6, " +
		"$7, $8, $9, $10) RETURNING id")
	if err != nil {
		return
	}

	defer stmt.Close()

	err = stmt.QueryRow(t.Name, t.Desc, t.CategoryID, t.Level, t.Price,
		t.Shared, t.Flag, t.MaxSharePrice, t.MinSharePrice,
		t.Opened).Scan(&t.ID)
	if err != nil {
		return
	}

	return
}

// GetTasks get all tasks in tasks table
func GetTasks(db *sql.DB) (tasks []Task, err error) {

	rows, err := db.Query("SELECT id, name, description, category_id, " +
		"level, price, shared, flag, max_share_price, " +
		"min_share_price, opened FROM task")
	if err != nil {
		return
	}

	defer rows.Close()

	for rows.Next() {
		var t Task

		err = rows.Scan(&t.ID, &t.Name, &t.Desc, &t.CategoryID,
			&t.Level, &t.Price, &t.Shared, &t.Flag,
			&t.MaxSharePrice, &t.MinSharePrice, &t.Opened)
		if err != nil {
			return
		}

		tasks = append(tasks, t)
	}

	return
}

// SetOpened open or close task
func SetOpened(db *sql.DB, taskID int, opened bool) (err error) {

	stmt, err := db.Prepare("UPDATE task SET opened=$1 WHERE id=$2")
	if err != nil {
		return err
	}

	defer stmt.Close()

	_, err = stmt.Exec(opened, taskID)
	if err != nil {
		return err
	}

	return nil
}
