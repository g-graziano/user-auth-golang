package postgres

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/g-graziano/userland/helper"
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
	GetActiveUser(user *models.User) ([]*models.User, error)

	// Token
	CreateToken(token *models.UserToken) error
	DeleteToken(token *models.UserToken) error
	GetToken(token *models.UserToken) ([]*models.UserToken, error)

	//BackupCode
	CreateBackUpCode(code *models.BackupCodes) error
	GetBackUpCode(code *models.BackupCodes) ([]*models.BackupCodes, error)

	//ClientID
	GetClientID(code *models.ClientID) ([]*models.ClientID, error)
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
			&models.UserToken{},
			&models.BackupCodes{},
			&models.ClientID{},
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
		"nonactive",
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

	_, err := p.DB[0].Exec(`
		UPDATE USERS SET 
			email = $1, 
			fullname = $2,
			password = $3, 
			location = $4,
			bio = $5,
			web = $6,
			picture = $7,
			status = $8,
			tfa = $9,
			enabled_tfa_at = $10,
			updated_at = $11
		WHERE id=$12`,
		user.Email,
		user.Fullname,
		user.Password,
		user.Location,
		user.Bio,
		user.Web,
		user.Picture,
		user.Status,
		user.TFA,
		user.EnabledTfaAt,
		updatedAt,
		user.ID,
	)

	if err != nil {
		return err
	}

	return nil
}

func (p *postgres) GetUser(user *models.User) ([]*models.User, error) {
	if user.Email != "" && user.XID != "" {
		// Search other user
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
				status,
				tfa,
				enabled_tfa_at,
				created_at,
				updated_at 
			FROM USERS WHERE email = $1 and x_id != $2`, user.Email, user.XID)

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
				&user.Location,
				&user.Bio,
				&user.Web,
				&user.Picture,
				&user.Status,
				&user.TFA,
				&user.EnabledTfaAt,
				&user.CreatedAt,
				&user.UpdatedAt,
			); err != nil {
				return nil, err
			}

			allUser = append(allUser, user)
		}

		return allUser, nil
	} else if user.Email != "" {
		// Search user by email status != deleted
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
				status,
				tfa,
				enabled_tfa_at,
				created_at,
				updated_at 
			FROM USERS WHERE email = $1 and status != 'deleted'`, user.Email)

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
				&user.Location,
				&user.Bio,
				&user.Web,
				&user.Picture,
				&user.Status,
				&user.TFA,
				&user.EnabledTfaAt,
				&user.CreatedAt,
				&user.UpdatedAt,
			); err != nil {
				return nil, err
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
				location,
				bio,
				web,
				picture,
				status,
				tfa,
				enabled_tfa_at,
				created_at,
				updated_at 
			FROM USERS WHERE x_id = $1 and status != 'deleted'`, user.XID)

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
				&user.Location,
				&user.Bio,
				&user.Web,
				&user.Picture,
				&user.Status,
				&user.TFA,
				&user.EnabledTfaAt,
				&user.CreatedAt,
				&user.UpdatedAt,
			); err != nil {
				return nil, err
			}

			allUser = append(allUser, user)
		}

		return allUser, nil
	}

	return nil, nil
}

func (p *postgres) GetActiveUser(user *models.User) ([]*models.User, error) {
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
				status,
				tfa,
				enabled_tfa_at,
				created_at,
				updated_at 
			FROM USERS WHERE email = $1 and status = 'active'`, user.Email)

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
				&user.Location,
				&user.Bio,
				&user.Web,
				&user.Picture,
				&user.Status,
				&user.TFA,
				&user.EnabledTfaAt,
				&user.CreatedAt,
				&user.UpdatedAt,
			); err != nil {
				return nil, err
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
				location,
				bio,
				web,
				picture,
				status,
				tfa,
				enabled_tfa_at,
				created_at,
				updated_at 
			FROM USERS WHERE x_id = $1 and status = 'active'`, user.XID)

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
				&user.Location,
				&user.Bio,
				&user.Web,
				&user.Picture,
				&user.Status,
				&user.TFA,
				&user.EnabledTfaAt,
				&user.CreatedAt,
				&user.UpdatedAt,
			); err != nil {
				return nil, err
			}

			allUser = append(allUser, user)
		}

		return allUser, nil
	}

	return nil, nil
}

func (p *postgres) CreateToken(token *models.UserToken) error {
	_, err := p.DB[0].Exec(`
		INSERT INTO USER_TOKENS (
			token,
			user_id,
			status,
			token_type,
			refresh_token,
			created_at,
			updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		token.Token,
		token.UserID,
		"active",
		token.TokenType,
		token.RefreshToken,
		time.Now(),
		time.Now(),
	)

	if err != nil {
		return err
	}

	return nil
}

func (p *postgres) DeleteToken(token *models.UserToken) error {
	if token.RefreshToken != helper.NullStringFunc("", false) {
		updatedAt := time.Now()

		_, err := p.DB[0].Exec(`
			UPDATE USER_TOKENS SET 
				status = $1,
				updated_at = $2
			WHERE refresh_token = $3`,
			"nonactive",
			updatedAt,
			token.RefreshToken,
		)

		if err != nil {
			return err
		}
	} else if token.Token != "" && token.UserID != 0 {
		updatedAt := time.Now()

		_, err := p.DB[0].Exec(`
			UPDATE USER_TOKENS SET
				status = $1,
				updated_at = $2
			WHERE token != $3 and user_id = $4`,
			"nonactive",
			updatedAt,
			token.Token,
			token.UserID,
		)

		if err != nil {
			return err
		}
	} else if token.Token != "" {
		updatedAt := time.Now()

		_, err := p.DB[0].Exec(`
			UPDATE USER_TOKENS SET
				status = $1,
				updated_at = $2
			WHERE token = $3`,
			"nonactive",
			updatedAt,
			token.Token,
		)

		if err != nil {
			return err
		}
	}

	return nil
}

func (p *postgres) GetToken(token *models.UserToken) ([]*models.UserToken, error) {
	var results []*models.UserToken

	if token.Token != "" {
		rows, err := p.DB[0].Query(`
				SELECT
					user_id,
					token,
					token_type,
					refresh_token,
					status,
					created_at,
					updated_at
				FROM USER_TOKENS WHERE 
					token = $1`, token.Token)

		if err != nil {
			return nil, err
		}

		defer rows.Close()

		for rows.Next() {
			var userToken = &models.UserToken{}
			if err := rows.Scan(
				&userToken.UserID,
				&userToken.Token,
				&userToken.TokenType,
				&userToken.RefreshToken,
				&userToken.Status,
				&userToken.CreatedAt,
				&userToken.UpdatedAt,
			); err != nil {
				return nil, err
			}

			results = append(results, userToken)
		}
	}

	return results, nil
}

func (p *postgres) CreateBackUpCode(code *models.BackupCodes) error {
	_, err := p.DB[0].Exec(`
		INSERT INTO BACKUP_CODES (
			user_id,
			codes
		) VALUES ($1, $2)`,
		code.UserID,
		code.Codes,
	)

	if err != nil {
		return err
	}

	return nil
}

func (p *postgres) GetBackUpCode(codes *models.BackupCodes) ([]*models.BackupCodes, error) {
	var results []*models.BackupCodes

	if codes.UserID != 0 && codes.Codes != "" {
		rows, err := p.DB[0].Query(`
				SELECT
					user_id,
					codes
				FROM BACKUP_CODES WHERE 
					user_id = $1 and codes = $2`, codes.UserID, codes.Codes)

		if err != nil {
			return nil, err
		}

		defer rows.Close()

		for rows.Next() {
			var userCode = &models.BackupCodes{}
			if err := rows.Scan(
				&userCode.UserID,
				&userCode.Codes,
			); err != nil {
				return nil, err
			}

			results = append(results, userCode)
		}
	}

	return results, nil
}

func (p *postgres) GetClientID(client *models.ClientID) ([]*models.ClientID, error) {
	var results []*models.ClientID

	if client.API != "" {
		rows, err := p.DB[0].Query(`
				SELECT
					api,
					name
				FROM CLIENT_IDS WHERE 
					api = $1`, client.API)

		if err != nil {
			return nil, err
		}

		defer rows.Close()

		for rows.Next() {
			var clienIDRow = &models.ClientID{}
			if err := rows.Scan(
				&clienIDRow.API,
				&clienIDRow.Name,
			); err != nil {
				return nil, err
			}

			results = append(results, clienIDRow)
		}
	}

	return results, nil
}
