package database

import (
	"os"
	"path/filepath"
	"time"

	customErrors "github.com/neoSnakex34/WasaPhoto/service/custom-errors"
	"github.com/neoSnakex34/WasaPhoto/service/structs"
)

const Folder string = "/tmp/wasaphoto/photofiles/"

// generate the identifier for the photo
// save the photofile path in the database
// save the photo in the database and create a new photo struct
func (db *appdbimpl) UploadPhoto(file []byte, upoloaderUserId structs.Identifier, format string) (structs.Photo, error) {

	var isValidId bool = false
	var newPhotoId structs.Identifier
	var err error
	var completePhotoPath string
	var uploaderId string = upoloaderUserId.Id
	// generate a new photo valid id
	for !isValidId && err == nil {

		newPhotoId, err = generateIdentifier("P")
		if err != nil {
			return structs.Photo{}, err
		}

		isValidId, err = db.validId(newPhotoId.Id, "P")

	}

	if err != nil {
		return structs.Photo{}, err
	}

	completePhotoPath = Folder + uploaderId + "/" + newPhotoId.Id + "." + format //

	// FIRST save the photo file in the filesystem
	err = savePhotoFile(file, completePhotoPath)
	if err != nil {
		return structs.Photo{}, err
	}

	date := time.Now().UTC().Format(time.RFC3339)

	// SECONDLY create the photo struct
	newPhoto := structs.Photo{
		PhotoId:            newPhotoId,
		UploaderUserId:     upoloaderUserId,
		LikeCounter:        0,
		Comments:           []structs.Comment{},
		CommentsCounter:    0,
		LikedByCurrentUser: false,
		Date:               date,
		PhotoPath:          uploaderId + "/" + newPhotoId.Id + "." + format, // needs to be partial for frontend
	}

	// AFTER FIRST TWO STEPS insert photo in the database
	err = db.insertPhotoInTable(newPhotoId.Id, upoloaderUserId.Id, date, completePhotoPath)
	if err != nil {
		return structs.Photo{}, err
	}

	return newPhoto, err // or nil directly
}

func (db *appdbimpl) RemovePhoto(photoId structs.Identifier, userId structs.Identifier) error {
	removedPhotoId := photoId.Id
	removerUserId := userId.Id

	approximatePhotoPath := Folder + removerUserId + "/" + removedPhotoId + ".*" // removes agnostically the image (without format) since ids are unique
	matchingPhoto, err := filepath.Glob(approximatePhotoPath)
	if err != nil {
		return err
	}

	if len(matchingPhoto) == 0 {
		return customErrors.ErrPhotoDoesNotExist
	}

	if len(matchingPhoto) > 1 {
		return customErrors.ErrCriticDuplicatedId
	}

	photoPath := matchingPhoto[0]

	err = db.removePhotoFromTable(removedPhotoId)
	if err != nil {
		return err
	}
	err = deletePhotoFile(photoPath)
	if err != nil {
		return err
	}
	return nil
}

// path will be the final path
func savePhotoFile(file []byte, path string) error {
	// retrieve the directory
	dir := filepath.Dir(path)
	// build the directory if it does not exist (it doesn't first time cause there will be a directory for every user)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}

	err = os.WriteFile(path, file, 0644) // permission may fail on linux but since this is distributed via docker it should be fine
	return err
}

func deletePhotoFile(path string) error {
	err := os.Remove(path)
	return err
}

func (db *appdbimpl) removePhotoFromTable(photoId string) error {
	_, err := db.c.Exec(`DELETE FROM photos WHERE photoId = ?`, photoId)
	return err
}

func (db *appdbimpl) insertPhotoInTable(photoId string, userId string, date string, path string) error {
	_, err := db.c.Exec(`INSERT INTO photos (photoId, userId, photoPath, date) VALUES (?, ?, ?, ?)`, photoId, userId, path, date)
	return err
}

func (db *appdbimpl) getStreamOfPhotos(followerIdsForUser []string, plainRequestorId string) ([]structs.Photo, error) {
	// for each follower i should retrieve a (photo slice) in order to build the stream
	// since i will need to sort the stream by date, i should return a complex struct instead of []string
	// and the access datas

	var stream []structs.Photo
	var tmpPhotos [][]structs.Photo
	for _, followerId := range followerIdsForUser {
		photos, err := db.getPhotosByUploaderId(followerId, plainRequestorId)
		if err != nil {
			return nil, err
		}
		tmpPhotos = append(tmpPhotos, photos)
	}

	// now stream will be an unsorted plain list of photos (no list of lists)
	for _, tmpList := range tmpPhotos {
		stream = append(stream, tmpList...)
	}

	return stream, nil

}
