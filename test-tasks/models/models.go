package models

import (
	"fmt"
	"regexp"
	"time"

	"github.com/go-playground/validator/v10"
)

var (
	validate   = validator.New()
	// Более надежное регулярное выражение для email
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
)

// User представляет пользователя системы
type User struct {
	ID        int       `json:"id" validate:"omitempty"`
	Email     string    `json:"email" validate:"required,email,max=255"`
	Password  string    `json:"-" validate:"omitempty"` // Не показываем в JSON
	Username  string    `json:"username" validate:"required,min=3,max=50"`
	CreatedAt time.Time `json:"created_at" validate:"omitempty"`
}

// Validate проверяет данные пользователя
func (u *User) Validate() error {
	if err := validate.Struct(u); err != nil {
		return err
	}
	if !emailRegex.MatchString(u.Email) {
		return fmt.Errorf("неверный формат email")
	}
	return nil
}

// Post представляет пост в блоге
type Post struct {
	ID        int       `json:"id" validate:"omitempty"`
	Title     string    `json:"title" validate:"required,max=200"`
	Content   string    `json:"content" validate:"required"`
	AuthorID  int       `json:"author_id" validate:"required"`
	Author    *User     `json:"author,omitempty" validate:"omitempty"`
	CreatedAt time.Time `json:"created_at" validate:"omitempty"`
	UpdatedAt time.Time `json:"updated_at" validate:"omitempty"`
}

// Validate проверяет данные поста
func (p *Post) Validate() error {
	return validate.Struct(p)
}

// Comment представляет комментарий к посту
type Comment struct {
	ID        int       `json:"id" validate:"omitempty"`
	PostID    int       `json:"post_id" validate:"required"`
	AuthorID  int       `json:"author_id" validate:"required"`
	Author    *User     `json:"author,omitempty" validate:"omitempty"`
	Content   string    `json:"content" validate:"required"`
	CreatedAt time.Time `json:"created_at" validate:"omitempty"`
}

// Validate проверяет данные комментария
func (c *Comment) Validate() error {
	return validate.Struct(c)
}
