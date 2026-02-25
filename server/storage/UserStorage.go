package storage

import (
	"database/sql"
	"errors"
	"server/models"
	"strconv"
)

func CreateUser(db *sql.DB, req models.PostUser) (models.User, error) {
	const q = `
		INSERT INTO users (name, age, department)
		VALUES ($1, $2, $3)
		RETURNING id;
	`

	var id int64
	err := db.QueryRow(q, req.Name, req.Age, req.Department).Scan(&id)
	if err != nil {
		return models.User{}, err
	}

	return models.User{
		Id:         id,
		Name:       req.Name,
		Age:        req.Age,
		Department: req.Department,
	}, nil
}

func GetUserByID(db *sql.DB, idStr string) (models.User, error) {
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		return models.User{}, errors.New("invalid id")
	}

	const q = `
		SELECT id, name, age, department
		FROM users
		WHERE id = $1;
	`

	var u models.User
	err = db.QueryRow(q, id).Scan(&u.Id, &u.Name, &u.Age, &u.Department)
	if err != nil {
		return models.User{}, err
	}
	return u, nil
}
