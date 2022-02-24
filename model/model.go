package model

import (
	"sync"
	"time"
)

type Chat struct {
	MessageID uint64 `json:"-"`

	ID   string   `json:"id"`
	Type ChatType `json:"type"`

	Meta []*Meta `json:"meta"`

	Image *string `json:"image"`
	Name  string  `json:"name"`

	Creator    *User   `json:"-"`
	OwnerLogin *string `json:"owner"`

	CreatedAt time.Time `json:"createdAt"`

	AllMembers  []*User    `json:"members"`
	AllMessages []*Message `json:"messages"`

	AllMembersByLogin map[string]*User    `json:"-"`
	AllMessagesByID   map[string]*Message `json:"-"`

	Observers map[string]*ChatObserver `json:"-"`

	M sync.RWMutex `json:"-"`
}

func (c *Chat) Owner() *User {
	if c.Creator == nil {
		return AnonymousUser
	}

	return c.Creator
}

func (c *Chat) Members(offset, first *int) []*User {
	begin := *offset
	count := *first

	if begin >= len(c.AllMembers) {
		return nil
	}

	if begin < 0 {
		count += begin
		begin = 0
	}

	if count < 0 {
		return nil
	}

	if len(c.AllMembers) < begin+count {
		count = len(c.AllMembers) - begin
	}

	return c.AllMembers[begin : begin+count]
}

func (c *Chat) Messages(offset, first *int) []*Message {
	c.M.RLock()
	defer c.M.RUnlock()

	begin := *offset
	count := *first

	if begin >= len(c.AllMessages) {
		return nil
	}

	if begin < 0 {
		count += begin
		begin = 0
	}

	if count < 0 {
		return nil
	}

	if len(c.AllMessages) < begin+count {
		count = len(c.AllMessages) - begin
	}

	messages := make([]*Message, 0, count)
	start := len(c.AllMessages) - begin - 1

	for i := start; i > start-count; i-- {
		messages = append(messages, c.AllMessages[i])
	}

	return messages
}

type ChatObserver struct {
	Login   string
	Message chan *Message
}

type Message struct {
	ID string `json:"id"`

	CreatedAt time.Time `json:"createdAt"`

	Author      *User   `json:"-"`
	AuthorLogin *string `json:"author"`

	Meta []*Meta `json:"meta"`
	Text string  `json:"text"`
}

func (m *Message) CreatedBy() *User {
	if m.Author == nil {
		return AnonymousUser
	}

	return m.Author
}

type Meta struct {
	Key string `json:"key"`
	Val string `json:"val"`
}

type User struct {
	Meta []*Meta `json:"meta"`

	Image *string `json:"image"`
	Login string  `json:"login"`
	Name  string  `json:"name"`

	Password string `json:"password"`

	M sync.RWMutex `json:"-"`
}
