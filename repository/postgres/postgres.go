package postgres

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/g-graziano/userland/models"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

type postgres struct {
	// gorms []*gorm.DB
	DB []*sql.DB
}

type Postgres interface {
	// User
	CreateUser(user *models.User) error
	UpdateUser(user *models.User) error
	GetUser(user *models.User) ([]*models.User, error)
}

func New(conn ...string) Postgres {

	// var gorms []*gorm.DB
	var DBS []*sql.DB
	for _, eachConn := range conn {
		DB, err := sql.Open("postgres", eachConn)

		defer func() {
			if r := recover(); r != nil {
				fmt.Println("msg:", r)
				defer DB.Close()
			}
		}()

		if err != nil {
			panic(err)
		}

		err = DB.Ping()

		if err != nil {
			panic(err)
		}

		fmt.Println("Database connected!")
		DBS = append(DBS, DB)

		// GORM for auto migrate database
		dbConn, err := gorm.Open("postgres", eachConn)
		if err != nil && dbConn == nil {
			fmt.Println(1, err, eachConn)
			continue
		}
		err = dbConn.DB().Ping()
		if err != nil {
			fmt.Println(2, err)
			continue
		}

		dbConn.Debug().AutoMigrate(
			&models.User{},
		)

		// gorms = append(gorms, dbConn)
	}
	// return &postgres{gorms: gorms, DB: DBS}
	return &postgres{DB: DBS}
}

func (p *postgres) CreateUser(user *models.User) error {
	_, err := p.DB[0].Exec(`
		INSERT INTO USERS (
			x_id,
			fullname, 
			email, 
			password,
			status,
			created_at,
			updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		user.XID,
		user.Fullname,
		user.Email,
		user.Password,
		1,
		time.Now(),
		time.Now(),
	)

	if err != nil {
		return err
	}

	return nil
}
func (p *postgres) UpdateUser(user *models.User) error {
	updatedAt := time.Now()

	p.DB[0].Exec(`
		UPDATE USERS SET 
			email = $1, 
			fullname = $2
			password = $3, 
			location = $4,
			bio = $5,
			web = $6,
			picture = $7,
			token = $8,
			status = $9,
			updated_at = $10
		WHERE id=$11`,
		user.Email,
		user.Fullname,
		user.Password,
		user.Location,
		user.Bio,
		user.Web,
		user.Picture,
		user.Token,
		user.Status,
		updatedAt,
		user.ID,
	)

	return nil
}
func (p *postgres) GetUser(user *models.User) ([]*models.User, error) {
	if user.Email != "" {
		var allUser []*models.User

		rows, err := p.DB[0].Query(`
			SELECT 
				id, 
				x_id,
				fullname,
				email,
				password,
				location,
				bio,
				web,
				picture,
				token,
				status,
				created_at,
				updated_at 
			FROM USERS WHERE email = $1`, user.Email)

		if err != nil {
			return nil, err
		}

		defer rows.Close()

		for rows.Next() {
			var user = &models.User{}
			if err := rows.Scan(
				&user.ID,
				&user.XID,
				&user.Fullname,
				&user.Email,
				&user.Password,
				&user.Token,
				&user.Status,
				&user.CreatedAt,
				&user.UpdatedAt,
			); err != nil {
				log.Fatal(err)
			}

			allUser = append(allUser, user)
		}

		return allUser, nil

	} else if user.XID != "" {
		var allUser []*models.User

		rows, err := p.DB[0].Query(`
			SELECT 
				id, 
				x_id,
				fullname,
				email,
				password,
				token,
				status,
				created_at,
				updated_at 
			FROM USERS WHERE x_id = $1`, user.XID)

		if err != nil {
			return nil, err
		}

		defer rows.Close()

		for rows.Next() {
			var user = &models.User{}
			if err := rows.Scan(
				&user.ID,
				&user.XID,
				&user.Fullname,
				&user.Email,
				&user.Password,
				&user.Token,
				&user.Status,
				&user.CreatedAt,
				&user.UpdatedAt,
			); err != nil {
				log.Fatal(err)
			}

			allUser = append(allUser, user)
		}

		return allUser, nil
	}

	return nil, nil
}
