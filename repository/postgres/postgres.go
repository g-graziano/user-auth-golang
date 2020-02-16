package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
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
	CreateToken(ctx context.Context, token *models.UserToken) error
	DeleteToken(token *models.UserToken) error
	GetToken(token *models.UserToken) ([]*models.UserToken, error)

	GetSession(token *models.UserToken) ([]*models.UserToken, error)

	//BackupCode
	CreateBackUpCode(code *models.BackupCodes) error
	GetBackUpCode(code *models.BackupCodes) ([]*models.BackupCodes, error)

	//ClientID
	GetClientID(code *models.ClientID) ([]*models.ClientID, error)

	//Event
	CreateEvent(ctx context.Context, event string, userID uint64) error
	GetEvent(user *models.User) ([]*models.Event, error)
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
			&models.Event{},
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
	var allUser []*models.User
	var rows *sql.Rows
	var err error

	if user.Email != "" && user.XID != "" {
		// Search other user
		rows, err = p.DB[0].Query(`
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
	} else if user.Email != "" {
		// Search user by email status != deleted
		rows, err = p.DB[0].Query(`
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
	} else if user.XID != "" {
		rows, err = p.DB[0].Query(`
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
	}

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

func (p *postgres) GetActiveUser(user *models.User) ([]*models.User, error) {
	var allUser []*models.User
	var rows *sql.Rows
	var err error

	if user.Email != "" {
		rows, err = p.DB[0].Query(`
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
	} else if user.XID != "" {
		rows, err = p.DB[0].Query(`
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
	}

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

func (p *postgres) CreateToken(ctx context.Context, token *models.UserToken) error {
	ipAddress := fmt.Sprintf("%v", ctx.Value(helper.StringToInterface("ip-address")))
	clientID, err := strconv.ParseUint(fmt.Sprintf("%v", ctx.Value(helper.StringToInterface("client-id"))), 0, 64)

	if err != nil {
		return err
	}

	_, err = p.DB[0].Exec(`
		INSERT INTO USER_TOKENS (
			token,
			user_id,
			status,
			token_type,
			refresh_token,
			ip_address,
			client_id,
			created_at,
			updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		token.Token,
		token.UserID,
		"active",
		token.TokenType,
		token.RefreshToken,
		ipAddress,
		clientID,
		time.Now(),
		time.Now(),
	)

	if err != nil {
		return err
	}

	return nil
}

func (p *postgres) DeleteToken(token *models.UserToken) error {
	var err error
	updatedAt := time.Now()
	if token.RefreshToken != helper.NullStringFunc("", false) {
		_, err = p.DB[0].Exec(`
			UPDATE USER_TOKENS SET 
				status = $1,
				updated_at = $2
			WHERE refresh_token = $3`,
			"nonactive",
			updatedAt,
			token.RefreshToken,
		)
	} else if token.Token != "" && token.UserID != 0 {
		_, err = p.DB[0].Exec(`
			UPDATE USER_TOKENS SET
				status = $1,
				updated_at = $2
			WHERE token != $3 and user_id = $4`,
			"nonactive",
			updatedAt,
			token.Token,
			token.UserID,
		)
	} else if token.Token != "" {
		_, err = p.DB[0].Exec(`
			UPDATE USER_TOKENS SET
				status = $1,
				updated_at = $2
			WHERE token = $3`,
			"nonactive",
			updatedAt,
			token.Token,
		)
	}

	if err != nil {
		return err
	}

	return nil
}

func (p *postgres) GetToken(token *models.UserToken) ([]*models.UserToken, error) {
	var results []*models.UserToken
	var rows *sql.Rows
	var err error

	if token.Token != "" {
		rows, err = p.DB[0].Query(`
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
	} else if token.UserID != 0 {
		rows, err = p.DB[0].Query(`
				SELECT
					user_id,
					token,
					token_type,
					refresh_token,
					status,
					created_at,
					updated_at
				FROM USER_TOKENS WHERE 
					user_id = $1`, token.UserID)
	}

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

	return results, nil
}

func (p *postgres) GetSession(token *models.UserToken) ([]*models.UserToken, error) {
	var results []*models.UserToken

	if token.UserID != 0 {
		rows, err := p.DB[0].Query(`
				SELECT
					u.user_id,
					u.token,
					u.token_type,
					u.refresh_token,
					u.status,
					u.client_id,
					c.name,
					u.ip_address,
					u.created_at,
					u.updated_at
				FROM USER_TOKENS as u
				LEFT JOIN CLIENT_IDS AS c ON u.client_id = c.id
				WHERE 
					u.user_id = $1 and u.status = 'active'`, token.UserID)

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
				&userToken.ClientID,
				&userToken.ClientName,
				&userToken.IPAddress,
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
					id,
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
				&clienIDRow.ID,
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

func (p *postgres) CreateEvent(ctx context.Context, event string, userID uint64) error {
	ipAddress := fmt.Sprintf("%v", ctx.Value(helper.StringToInterface("ip-address")))
	userAgent := fmt.Sprintf("%v", ctx.Value(helper.StringToInterface("user-agent")))
	clientID, err := strconv.ParseUint(
		fmt.Sprintf("%v", ctx.Value(helper.StringToInterface("client-id"))),
		0,
		64,
	)

	if err != nil {
		return err
	}

	_, err = p.DB[0].Exec(`
		INSERT INTO EVENTS (
			user_id,
			event, 
			ua, 
			ip_address,
			client_id,
			created_at,
			updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		userID,
		event,
		userAgent,
		ipAddress,
		clientID,
		time.Now(),
		time.Now(),
	)

	if err != nil {
		return err
	}

	return nil
}

func (p *postgres) GetEvent(user *models.User) ([]*models.Event, error) {
	var results []*models.Event

	if user.ID != 0 {
		rows, err := p.DB[0].Query(`
				SELECT
					u.user_id,
					u.event,
					u.ua,
					u.client_id,
					c.name,
					u.ip_address,
					u.created_at
				FROM EVENTS as u
				LEFT JOIN CLIENT_IDS AS c ON u.client_id = c.id
				WHERE 
					u.user_id = $1`, user.ID)

		if err != nil {
			return nil, err
		}

		defer rows.Close()

		for rows.Next() {
			var event = &models.Event{}
			if err := rows.Scan(
				&event.UserID,
				&event.Event,
				&event.UA,
				&event.ClientID,
				&event.ClientName,
				&event.IPAddress,
				&event.CreatedAt,
			); err != nil {
				return nil, err
			}

			results = append(results, event)
		}
	}

	return results, nil
}
