package db

import (
	"database/sql"
	"fmt"
)


type Task struct {
	ID      string `json:"id" db:"id"`
	Date    string `json:"date" db:"date"`
	Title   string `json:"title" db:"title"`
	Comment string `json:"comment" db:"comment"`
	Repeat  string `json:"repeat" db:"repeat"`
}

func AddTask(t *Task) (string, error) {
	res, err := DB.Exec(`INSERT INTO scheduler(date,title,comment,repeat) VALUES(?,?,?,?)`, t.Date, t.Title, t.Comment, t.Repeat)
	if err != nil {
		return "", fmt.Errorf("insert task: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return "", fmt.Errorf("lastInsertId: %w", err)
	}
	return fmt.Sprintf("%d", id), nil
}

func UpdateTask(t *Task) error {
	res, err := DB.Exec(
		`UPDATE scheduler SET date=?, title=?, comment=?, repeat=? WHERE id=?`,
		t.Date, t.Title, t.Comment, t.Repeat, t.ID,
	)
	if err != nil {
		return fmt.Errorf("update task: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rowsAffected: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("incorrect ID for updating task")
	}
	return nil
}

func UpdateDate(id, date string) error {
	res, err := DB.Exec(`UPDATE scheduler SET date=? WHERE id=?`, date, id)
	if err != nil {
		return fmt.Errorf("update date: %w", err)
	}
	a, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rowsAffected: %w", err)
	}
	if a == 0 {
		return fmt.Errorf("incorrect ID for updating task")
	}
	return nil
}

func GetTask(id string) (*Task, error) {
	row := DB.QueryRow(`SELECT id,date,title,comment,repeat FROM scheduler WHERE id=?`, id)
	var t Task
	err := row.Scan(&t.ID, &t.Date, &t.Title, &t.Comment, &t.Repeat)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("task not found")
		}
		return nil, fmt.Errorf("scan task: %w", err)
	}
	return &t, nil
}

func DeleteTask(id string) error {
	res, err := DB.Exec(`DELETE FROM scheduler WHERE id=?`, id)
	if err != nil {
		return fmt.Errorf("delete task: %w", err)
	}
	a, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rowsAffected: %w", err)
	}
	if a == 0 {
		return fmt.Errorf("incorrect ID for deleting task")
	}
	return nil
}

func GetTasks(limit int, search string) ([]*Task, error) {
	var rows *sql.Rows
	var err error

	if search == "" {
		rows, err = DB.Query(
			`SELECT id,date,title,comment,repeat FROM scheduler ORDER BY date LIMIT ?`,
			limit,
		)
	} else {
		if len(search) == 10 && search[2] == '.' && search[5] == '.' {
			yyyy := search[6:10]
			mm := search[3:5]
			dd := search[0:2]
			rows, err = DB.Query(
				`SELECT id,date,title,comment,repeat FROM scheduler WHERE date=? ORDER BY date LIMIT ?`,
				yyyy+mm+dd, limit,
			)
		} else {
			like := "%" + search + "%"
			rows, err = DB.Query(
				`SELECT id,date,title,comment,repeat FROM scheduler
				 WHERE title LIKE ? OR comment LIKE ?
				 ORDER BY date LIMIT ?`,
				like, like, limit,
			)
		}
	}
	if err != nil {
		return nil, fmt.Errorf("query tasks: %w", err)
	}
	defer rows.Close()

	var res []*Task
	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.ID, &t.Date, &t.Title, &t.Comment, &t.Repeat); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		res = append(res, &t)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration: %w", err)
	}

	return res, nil
}
