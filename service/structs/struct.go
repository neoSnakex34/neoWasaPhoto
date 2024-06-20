package structs

type Identifier struct {
	Id string `json:"identifier"`
}

type UserProfile struct {
	UserId           Identifier `json:"userId"`
	Username         string     `json:"username"`
	FollowerCounter  int        `json:"followersCounter"`
	FollowingCounter int        `json:"followingCounter"`
	PhotoCounter     int        `json:"photoCounter"`
	Photos           []Photo    `json:"photos"` // list of photos

}

type User struct {
	UserId   Identifier `json:"userId"`
	Username string     `json:"username"`
}

type UserFromQuery struct {
	User                 User `json:"user"`
	IsRequestorBanned    bool `json:"isRequestorBanned"`
	RequestorHasBanned   bool `json:"requestorHasBanned"`
	RequestorHasFollowed bool `json:"requestorHasFollowed"`
}

type Photo struct {
	PhotoId            Identifier `json:"photoId"`
	UploaderUserId     Identifier `json:"uploaderUserId"`
	UploaderUsername   string     `json:"uploaderUsername"`
	LikeCounter        int        `json:"likeCounter"`
	Comments           []Comment  `json:"comments"`
	CommentsCounter    int        `json:"commentsCounter"`
	LikedByCurrentUser bool       `json:"likedByCurrentUser"`
	Date               string     `json:"date"`
	PhotoPath          string     `json:"photoPath"` // in openapi this is represented as photofile
}

type Comment struct {
	CommentId          Identifier `json:"commentId"`
	CommentingUserId   Identifier `json:"commentingUserId"`   // commenter id
	CommentingUsername string     `json:"commentingUsername"` // commenter username not saved directly in db but obtained via db utilities
	PhotoId            Identifier `json:"photoId"`
	Body               string     `json:"commentBody"`
	Date               string     `json:"commentDate"`
}

// BodyRequest ... request body for a comment
type BodyRequest struct {
	Body string `json:"body"`
}
