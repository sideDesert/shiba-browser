package store

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"

	"sideDesert/shiba/internal/server/dto"
	"sideDesert/shiba/internal/server/lib"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	pool *pgxpool.Pool
}

func NewStore(ctx context.Context, dbUrl string) (*Store, error) {
	config, err := pgxpool.ParseConfig(dbUrl)
	if err != nil {
		log.Printf("Unable to create config using dbUrl: %v", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		log.Printf("Unable to create pool using config: %v", err)
		return nil, err
	}

	return &Store{
		pool: pool,
	}, nil
}

func (s *Store) Close(ctx context.Context) error {
	s.pool.Close()

	return nil
}

func (s *Store) CreateUser(ctx context.Context, user *dto.SignupUserRequest) (*string, error) {
	query := `INSERT INTO users (name, email, username, password_hash)
			  VALUES ($1, $2, $3, $4)
			  RETURNING user_id;`

	hashedPassword, err := lib.HashPassword(user.Password)
	if err != nil {
		log.Println("Error in CreateUser[hashPassword]", err)
		return nil, err
	}

	var userID string
	err = s.pool.QueryRow(ctx, query, user.Name, user.Email, user.Username, hashedPassword).Scan(&userID)
	if err != nil {
		log.Println("Error in CreateUser[QueryRow.Scan]:", err)
		return nil, err
	}

	return &userID, nil
}

func (s *Store) GetUserByEmail(ctx context.Context, email string) (*lib.User, error) {
	query := `SELECT id, user_id, name, email, password_hash, username, created_at, status, profile_picture FROM users WHERE email = $1`
	row := s.pool.QueryRow(ctx, query, email)
	user := lib.User{}
	err := row.Scan(
		&user.Id,
		&user.UserId,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&user.Username,
		&user.CreatedAt,
		&user.Status,
		&user.ProfilePicture,
	)

	if err != nil {
		log.Println("Error in GetUserByEmail[row.Scan]", err)
		return nil, err
	}
	return &user, nil
}

func (s *Store) GetUserById(ctx context.Context, user_id string) (*lib.User, error) {
	query := `SELECT id, user_id, name, email, password_hash, username, created_at, status, profile_picture FROM users WHERE user_id = $1`
	row := s.pool.QueryRow(ctx, query, user_id)
	user := lib.User{}
	err := row.Scan(
		&user.Id,
		&user.UserId,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&user.Username,
		&user.CreatedAt,
		&user.Status,
		&user.ProfilePicture,
	)

	if err != nil {
		log.Println("Error in GetUserById[row.Scan]", err)
		return nil, err
	}
	return &user, nil
}

// TODO: Complete this
func (s *Store) GetOrCreateUser(ctx context.Context, oauthUser *dto.OAuthGoogleUserDataResponse) (*lib.User, bool, error) {
	email := oauthUser.Email
	log.Println("Email", email)
	// Check if user exists
	user, err := s.GetUserByEmail(ctx, email)
	newUser := false

	if err != nil {
		log.Println("GetOrCreateUser: Creating New User")
		newUser = true
		dummyPass, err := lib.GenerateSecureRandomID(12)
		if err != nil {
			log.Println("Error in GetOrCreateUser[CreateUser]", err)
			return nil, newUser, err
		}

		hashedPassword, err := lib.HashPassword(dummyPass)
		if err != nil {
			log.Println("Error in GetOrCreateUser[CreateUser]", err)
			return nil, newUser, err
		}

		_, err = s.CreateUser(ctx, &dto.SignupUserRequest{
			Name:     oauthUser.Name,
			Username: strings.Split(oauthUser.Email, "@")[0],
			Email:    oauthUser.Email,
			Password: hashedPassword,
		})

		if err != nil {
			log.Println("Error in GetOrCreateUser[CreateUser]", err)
			return nil, newUser, err
		}

		newUser = true

	}
	return user, newUser, nil
}

func (s *Store) GetChatRoomById(ctx context.Context, chatroomId int32) (*lib.Chatroom, error) {
	q := "SELECT * FROM chatrooms WHERE id = $1"
	row := s.pool.QueryRow(ctx, q, chatroomId)
	chatRoom := lib.Chatroom{}
	err := row.Scan(&chatRoom.Id, &chatRoom.Name, &chatRoom.ProfilePicture, &chatRoom.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return &chatRoom, fmt.Errorf("GetChatRoomById: No chatroom with id %d", chatroomId)
		}
		return &chatRoom, fmt.Errorf("GetChatRoomById: %d: %s", chatroomId, err)
	}

	return &chatRoom, nil
}

func (s *Store) GetUsersByChatroomId(ctx context.Context, chatroomId string) ([]lib.UserId, error) {
	q := "SELECT user_id FROM user_chatrooms WHERE chatroom_id = $1"
	row, err := s.pool.Query(ctx, q, chatroomId)
	userIds := make([]lib.UserId, 0)
	if err == pgx.ErrNoRows {
		return userIds, nil
	}
	if err != nil {
		log.Println("Error in [store].GetUsersByChatroomId", err)
		return nil, err
	}
	defer row.Close()
	for row.Next() {
		var userId lib.UserId
		if err := row.Scan(&userId); err != nil {
			return nil, err
		}
		userIds = append(userIds, userId)
	}
	return userIds, nil
}

func (s *Store) CreateChatRoom(ctx context.Context, chatroom dto.CreateChatRoomRequest) (string, error) {
	q := "INSERT INTO chatrooms (name, profile_picture, direct_message) VALUES ($1, $2, $3) RETURNING id"
	row := s.pool.QueryRow(ctx, q, chatroom.Name, chatroom.ProfilePicture, chatroom.DirectMessage)
	var chatRoomId string
	if err := row.Scan(&chatRoomId); err != nil {
		return "", err
	}

	s.AddParticipantsToChatRoom(ctx, chatRoomId, chatroom.Participants)
	return chatRoomId, nil
}

func (s *Store) CreateChatRoomByName(ctx context.Context, name string) (string, error) {
	q := "INSERT INTO chatooms (name) VALUES ($1)"
	row := s.pool.QueryRow(ctx, q, name)
	var chatRoomId string
	if err := row.Scan(&chatRoomId); err != nil {
		return "", nil
	}
	return chatRoomId, nil
}

func (s *Store) AddParticipantsToChatRoom(ctx context.Context, chatroom_id string, participants []string) error {
	if len(participants) == 0 {
		log.Println("No participants provided")
		return nil
	}

	q := "INSERT INTO user_chatrooms (user_id, chatroom_id) VALUES ($1, $2)"

	for _, userId := range participants {
		commandTag, err := s.pool.Exec(ctx, q, userId, chatroom_id)
		if err != nil {
			return fmt.Errorf("Error in Adding participants: %s", err.Error())
		}
		rowsAffected := commandTag.RowsAffected()
		log.Println("Rows Affected:", rowsAffected)
	}
	return nil
}

func (s *Store) GetLast50ChatRoomMessages(ctx context.Context, chatroomId string, offset int) ([]lib.Message, error) {
	q := `SELECT m.id, u.name, m.sender, m.recipient, m.content, m.created_at
FROM messages m
LEFT JOIN users u ON u.user_id = m.sender
WHERE m.recipient = $1
ORDER BY m.created_at DESC
LIMIT 50 OFFSET $2;
`

	rows, err := s.pool.Query(ctx, q, chatroomId, 50*offset)
	messages := make([]lib.Message, 0)

	if err == pgx.ErrNoRows {
		return messages, nil
	}

	if err != nil {
		log.Println("Error in GetChatRoomHistory:", err.Error())
		return messages, err
	}

	for rows.Next() {
		tempMsg := lib.Message{}

		err := rows.Scan(&tempMsg.Id, &tempMsg.SenderName, &tempMsg.Sender, &tempMsg.ChatroomId, &tempMsg.Content, &tempMsg.CreatedAt)
		if err != nil {
			log.Println("Error in GetChatRoomHistory[Scan]:", err.Error())
			continue
		}

		messages = append(messages, tempMsg)
	}

	return messages, nil
}

type StoreChatMessageDto struct {
	Id         string `json:"id"`
	Sender     string `json:"string"`
	ChatroomId string `json:"chatroomId"`
	SenderName string `json:"sender_name"`
	Content    string `json:"content"`
	CreatedAt  string `json:"created_at"`
}

func (s *Store) StoreChatRoomMessage(ctx context.Context, msg StoreChatMessageDto) error {
	q := "INSERT INTO messages (sender, content, recipient) VALUES ($1, $2, $3)"

	log.Println("sender", msg.Sender)
	log.Println("content", msg.Content)
	log.Println("ChatroomId", msg.ChatroomId)

	_, err := s.pool.Exec(ctx, q, msg.Sender, msg.Content, msg.ChatroomId)

	if err != nil {
		log.Println("⁉️ Error in StoreChatRoomMessage:", err)
		return err
	}

	return nil
}

func (s *Store) GetChatRoomsByUserId(ctx context.Context, userId string) ([]lib.Chatroom, error) {
	q := `SELECT c.id, c.name, c.profile_picture, c.created_at, c.direct_message
	FROM chatrooms c
	JOIN user_chatrooms uc ON c.id = uc.chatroom_id
	WHERE uc.user_id = $1`

	rows, err := s.pool.Query(ctx, q, userId)
	chatroomList := make([]lib.Chatroom, 0)

	if err == pgx.ErrNoRows {
		return chatroomList, nil
	}
	if err != nil {
		log.Println("Error in GetChatRoomsByUserId:", err.Error())
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		c := lib.Chatroom{}
		rows.Scan(&c.Id, &c.Name, &c.ProfilePicture, &c.CreatedAt, &c.DirectMessage)
		chatroomList = append(chatroomList, c)
	}

	return chatroomList, rows.Err()
}

func (s *Store) SearchUsers(ctx context.Context, userId string, searchString string, limit int) ([]dto.SearchUserResponse, error) {
	q := `SELECT name, user_id, username, email, status, profile_picture FROM search_users($2, $3) WHERE user_id != $1`

	response := make([]dto.SearchUserResponse, 0)
	rows, err := s.pool.Query(ctx, q, userId, searchString, limit)
	if err == pgx.ErrNoRows {
		log.Println("NO Rows in Store.SearchUsers[Query]:")
		return response, nil
	}

	if err != nil {
		log.Println("Error in Store.SearchUsers[Query]:", err)
		return response, err
	}

	for rows.Next() {
		temp := dto.SearchUserResponse{}
		err := rows.Scan(&temp.Name, &temp.UserId, &temp.Username, &temp.Email, &temp.Status, &temp.ProfilePicture)
		if err != nil {
			log.Println("Error in Store.SearchUsers[Scan]:", err)
			continue
		}
		response = append(response, temp)
	}

	return response, nil
}

type UserFriend struct {
	UserId         string         `json:"user_id"`
	Name           string         `json:"name"`
	Username       string         `json:"username"`
	ProfilePicture sql.NullString `json:"profile_picture"`
	UserStatus     sql.NullString `json:"user_status"`
	FriendStatus   string         `json:"friend_status"`
	FriendshipId   string         `json:"request_id"`
	ChatroomId     string         `json:"chatroom_id"`
}

func (s *Store) GetFriendsByUserId(ctx context.Context, userId string) ([]UserFriend, error) {
	q := `SELECT
u.user_id,
u.name,
u.username,
u.profile_picture,
u.status,
f.status AS friendship_status,
f.id AS friendship_id,
uc.chatroom_id AS chatroom_id
FROM friends f
JOIN
    users u ON u.user_id = CASE
        WHEN f.user_id1 = $1 THEN f.user_id2
        ELSE f.user_id1
    END
LEFT JOIN
    user_chatrooms uc ON (uc.user_id = u.user_id AND uc.chatroom_id IN (
        SELECT chatroom_id
        FROM user_chatrooms
        WHERE user_id = $1
    ))
WHERE $1 IN (f.user_id1, f.user_id2)
AND f.status = 'accepted'`

	response := make([]UserFriend, 0)

	rows, err := s.pool.Query(ctx, q, userId)

	if err == pgx.ErrNoRows {
		return response, nil
	}

	if err != nil {
		log.Println("Error in Store.GetFriendsByUserId[Query]:", err)
		return response, err
	}

	for rows.Next() {
		t := UserFriend{}
		err := rows.Scan(&t.UserId, &t.Name, &t.Username, &t.ProfilePicture, &t.UserStatus, &t.FriendStatus, &t.FriendshipId, &t.ChatroomId)
		if err != nil {
			log.Println("Error in Store.GetFriendsByUserId[Scan]:", err)
			continue
		}
		response = append(response, t)
	}

	return response, nil
}

func (s *Store) GetFriendRelationsByUserId(ctx context.Context, userId string) ([]lib.FriendRelations, error) {
	q := "SELECT id, user_id1, user_id2, created_at, updated_at, status FROM friends WHERE user_id1 = $1 OR user_id2 = $1"
	response := make([]lib.FriendRelations, 0)

	rows, err := s.pool.Query(ctx, q, userId)
	if err == pgx.ErrNoRows {
		return response, nil
	}

	if err != nil {
		log.Println("Error in Store.GetFriendRelationsByUserId[Query]:", err)
		return response, err
	}

	for rows.Next() {
		f := lib.FriendRelations{}
		err := rows.Scan(&f.Id, &f.UserId1, &f.UserId2, &f.CreatedAt, &f.UpdatedAt, &f.Status)
		if err != nil {
			log.Println("Error in Store.GetFriendRelationsByUserId[Scan]:", err)
			continue
		}

		response = append(response, f)
	}

	return response, nil
}

func (s *Store) GetFriendRequestsByUserId(ctx context.Context, userId string) ([]dto.FriendRequestResponse, error) {
	q := `SELECT u.name, u.username, u.profile_picture, f.id, f.user_id1, f.created_at, f.status
	FROM friends f
	JOIN users u
	ON f.user_id2 = u.user_id
	WHERE f.user_id2 = $1 AND f.status = $2`

	response := make([]dto.FriendRequestResponse, 0)

	rows, err := s.pool.Query(ctx, q, userId, "pending")
	if err == pgx.ErrNoRows {
		return response, nil
	}

	if err != nil {
		log.Println("Error in Store.GetFriendRequestsByUserId[Query]:", err)
		return response, err
	}

	for rows.Next() {
		f := dto.FriendRequestResponse{}
		err := rows.Scan(&f.Name, &f.Username, &f.ProfilePicture, &f.RequestId, &f.UserId, &f.CreatedAt, &f.Status)
		if err != nil {
			log.Println("Error in Store.GetFriendRequstsByUserId[[Scan]:", err)
			continue
		}

		response = append(response, f)
	}

	return response, nil
}

func (s *Store) ChangeFriendStatus(ctx context.Context, id string, status string) error {
	q := "UPDATE friends SET status = $1 WHERE id = $2"
	log.Println("Chaging Friend Request Status for ID", id, "to", status)
	_, err := s.pool.Exec(ctx, q, status, id)
	if err != nil {
		log.Println("Error in Store.ChangeFriendStatus[Scan]:", err)
		return err
	}
	return nil
}

func (s *Store) InsertFriendStatus(ctx context.Context, userId string, friendId string, status string) error {
	q := "INSERT INTO friends (user_id1, user_id2, status) VALUES ($1, $2, $3)"
	fmt.Println("userId", userId, "friendId", friendId, status, "status")
	_, err := s.pool.Exec(ctx, q, userId, friendId, status)

	if err != nil {
		log.Println("Error in Store.InsertFriendStatus[Scan]:", err)
		return err
	}
	return nil
}

func (s *Store) GetRemoteByChatroomId(ctx context.Context, chatroomId string) (*dto.RemoteResponse, error) {
	q := `SELECT u.name, u.username, r.user_id, u.status
	FROM remote r
	JOIN users u
	ON r.user_id = u.user_id
	WHERE r.chatroom_id = $1`

	response := &dto.RemoteResponse{}

	rows, err := s.pool.Query(ctx, q, chatroomId)
	if err == pgx.ErrNoRows {
		return response, nil
	}

	if err != nil {
		log.Println("Error in Store.GetRemoteByChatroomId[Query]:", err)
		return response, err
	}

	for rows.Next() {
		err := rows.Scan(&response.Name, &response.Username, &response.UserId, &response.Status)
		if err != nil {
			log.Println("Error in Store.GetRemoteByChatroomId[Scan]:", err)
			continue
		}
	}

	return response, nil
}

func (s *Store) UpdateRemote(ctx context.Context, chatroomId string, userId string) error {
	q := "UPDATE remote SET user_id = $1 WHERE chatroom_id = $2"
	_, err := s.pool.Exec(ctx, q, userId, chatroomId)
	if err != nil {
		log.Println("Error in Store.CreateRemote[Exec]:", err)
		return err
	}

	return nil
}
func (s *Store) CreateRemote(ctx context.Context, chatroomId string, userId string) error {
	q := "INSERT INTO remote (user_id, chatroom_id) VALUES ($1, $2)"
	_, err := s.pool.Exec(ctx, q, userId, chatroomId)
	if err != nil {
		log.Println("Error in Store.CreateRemote[Exec]:", err)
		return err
	}

	return nil
}
