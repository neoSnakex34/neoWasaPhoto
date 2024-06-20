/*
Package database is the middleware between the app database and the code. All data (de)serialization (save/load) from a
persistent database are handled here. Database specific logic should never escape this package.

To use this package you need to apply migrations to the database if needed/wanted, connect to it (using the database
data source name from config), and then initialize an instance of AppDatabase from the DB connection.

For example, this code adds a parameter in `webapi` executable for the database data source name (add it to the
main.WebAPIConfiguration structure):

	DB struct {
		Filename string `conf:""`
	}

This is an example on how to migrate the DB and connect to it:

	// Start Database
	logger.Println("initializing database support")
	db, err := sql.Open("sqlite3", "./foo.db")
	if err != nil {
		logger.WithError(err).Error("error opening SQLite DB")
		return fmt.Errorf("opening SQLite: %w", err)
	}
	defer func() {
		logger.Debug("database stopping")
		_ = db.Close()
	}()

Then you can initialize the AppDatabase and pass it to the api package.
*/
package database

import (
	"database/sql"
	"errors"
	"log"
	"strings"

	"github.com/neoSnakex34/WasaPhoto/service/structs"
)

// AppDatabase is the high level interface for the DB
// methods are exported ones, hence they are written with capital first letter
type AppDatabase interface {
	DoLogin(username string) (structs.Identifier, error)                // done
	SetMyUserName(newUsername string, userId string, mode string) error // done

	GetUserList(requestorUserId structs.Identifier) ([]structs.UserFromQuery, error)                                  // done
	GetUserProfile(profileUserId structs.Identifier, requestorUserId structs.Identifier) (structs.UserProfile, error) // done
	GetMyStream(userId structs.Identifier) ([]structs.Photo, error)                                                   // decide wether to return list of paths or list of photos structs

	FollowUser(userId structs.Identifier, followedId structs.Identifier) error   // done
	UnfollowUser(userId structs.Identifier, followerId structs.Identifier) error // done

	BanUser(bannerId structs.Identifier, bannedId structs.Identifier) error           // done
	UnbanUser(bannerUserId structs.Identifier, bannedUserId structs.Identifier) error // done

	UploadPhoto(file []byte, uploaderUserId structs.Identifier, format string) (structs.Photo, error) // done
	RemovePhoto(photoId structs.Identifier, userId structs.Identifier) error                          // done

	CommentPhoto(commentedPhotoId structs.Identifier, requestorUserId structs.Identifier, body string) (structs.Comment, error) // done
	UncommentPhoto(commentId structs.Identifier) error                                                                          // done

	LikePhoto(userId structs.Identifier, photoId structs.Identifier) error   // done
	UnlikePhoto(userId structs.Identifier, photoId structs.Identifier) error // done

	Ping() error
}

type appdbimpl struct {
	c *sql.DB
}

// New returns a new instance of AppDatabase based on the SQLite connection `db`.
// `db` is required - an error will be returned if `db` is `nil`.
func New(db *sql.DB) (AppDatabase, error) {
	if db == nil {
		return nil, errors.New("database is required when building a AppDatabase")
	}

	// start creating the AppDatabase if needed

	var tableName string
	err := db.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name='users';`).Scan(&tableName)
	if errors.Is(err, sql.ErrNoRows) {

		userTable := `CREATE TABLE users (
			userId VARCHAR(11) NOT NULL PRIMARY KEY,
			username VARCHAR(18) NOT NULL UNIQUE

			)`

		followerTable := `CREATE TABLE followers (
			followerId VARCHAR(11) NOT NULL,
			followedId VARCHAR(11) NOT NULL,
			PRIMARY KEY (followerId, followedId),
			FOREIGN KEY (followerId) REFERENCES users(userId),
			FOREIGN KEY (followedId) REFERENCES users(userId)
		)`

		bansTable := `CREATE TABLE bans (
			bannerId VARCHAR(11) NOT NULL,
			bannedId VARCHAR(11) NOT NULL,
			PRIMARY KEY (bannerId, bannedId),
			FOREIGN KEY (bannerId) REFERENCES users(userId),
			FOREIGN KEY (bannedId) REFERENCES users(userId)
		)`

		photoTable := `CREATE TABLE photos (
			photoId VARCHAR(11) NOT NULL PRIMARY KEY,
			userId VARCHAR(11) NOT NULL,
			photoPath TEXT,
			date TEXT,
			FOREIGN KEY (userId) REFERENCES users(userId)
		)`

		likeTable := `CREATE TABLE likes (
			likerId VARCHAR(11) NOT NULL,
			photoId VARCHAR(11) NOT NULL,
			FOREIGN KEY (likerId) REFERENCES users(userId)
			FOREIGN KEY (photoId) REFERENCES photos(photoId) ON DELETE CASCADE
			PRIMARY KEY (likerId, photoId)
		)`

		commentTable := `CREATE TABLE comments (
			commentId VARCHAR(11) NOT NULL PRIMARY KEY,
			userId VARCHAR(11) NOT NULL,
			photoId VARCHAR(11) NOT NULL,
			body TEXT,
			date TEXT,
			FOREIGN KEY (userId) REFERENCES users(userId),
			FOREIGN KEY (photoId) REFERENCES photos(photoId) ON DELETE CASCADE
		)`

		err = runCreateQueries(db, userTable, followerTable, bansTable, photoTable, likeTable, commentTable)
		if err != nil {
			log.Println("Error creating tables")
		}
	}

	return &appdbimpl{
		c: db,
	}, nil
}

func runCreateQueries(db *sql.DB, queries ...string) error {
	// this is easy to scale but be wary not to use fmtSprintf for queries
	// to avoid sql injection vulnerabilities
	for _, query := range queries {
		log.Println("creating table: ", strings.Split(query, " ")[2])
		_, err := db.Exec(query)

		if err != nil {
			log.Println("error creating table: ", query, err)
			return err
		}
	}
	return nil
}

func (db *appdbimpl) Ping() error {
	return db.c.Ping()
}
