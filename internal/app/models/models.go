package models

import (
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/jwma/jump-jump/internal/app/db"
	"github.com/jwma/jump-jump/internal/app/utils"
	"log"
	"time"
)

type ShortLink struct {
	Id          string    `json:"id"`
	Url         string    `json:"url"`
	Description string    `json:"description"`
	IsEnable    bool      `json:"is_enable"`
	CreatedBy   string    `json:"created_by"`
	CreateTime  time.Time `json:"create_time"`
	UpdateTime  time.Time `json:"update_time"`
}

type UpdateShortLinkParameter struct {
	Url         string `json:"url" binding:"required"`
	Description string `json:"description"`
	IsEnable    bool   `json:"is_enable"`
}

func (s *ShortLink) key() string {
	return fmt.Sprintf("link:%s", s.Id)
}

func (s *ShortLink) GenerateId() error {
	client := db.GetRedisClient()

	for true {
		s.Id = utils.RandStringRunes(6)
		_, err := client.Get(s.key()).Result()
		if err == redis.Nil {
			return nil
		}
		if err != nil {
			log.Println(err)
			return err
		}
	}

	return nil
}

func (s *ShortLink) Save() error {
	if s.Id == "" {
		return fmt.Errorf("id错误")
	}
	if s.Url == "" {
		return fmt.Errorf("请填写url")
	}

	s.CreateTime = time.Now()
	s.UpdateTime = time.Now()
	c := db.GetRedisClient()
	j, _ := json.Marshal(s)
	c.Set(s.key(), string(j), 0)

	return nil
}

func (s *ShortLink) Get() error {
	if s.Id == "" {
		return fmt.Errorf("短链接不存在")
	}

	c := db.GetRedisClient()
	rs, err := c.Get(s.key()).Result()
	if err != nil {
		log.Printf("fail to get short link with key: %s, error: %v\n", s.key(), err)
		return fmt.Errorf("短链接不存在")
	}

	err = json.Unmarshal([]byte(rs), s)
	if err != nil {
		log.Printf("fail to unmarshal short link, key: %s, error: %v\n", s.key(), err)
		return fmt.Errorf("短链接不存在")
	}

	return nil
}

func (s *ShortLink) Update(u *UpdateShortLinkParameter) error {
	s.Url = u.Url
	s.Description = u.Description
	s.IsEnable = u.IsEnable
	s.UpdateTime = time.Now()

	return s.Save()
}

type RequestHistory struct {
	link *ShortLink `json:"-"`
	IP   string     `json:"ip"`
	UA   string     `json:"ua"`
	Time time.Time  `json:"time"`
}

func NewRequestHistory(link *ShortLink, IP string, UA string) *RequestHistory {
	return &RequestHistory{link: link, IP: IP, UA: UA}
}

func (r *RequestHistory) key() string {
	return fmt.Sprintf("history:%s", r.link.Id)
}

func (r *RequestHistory) Save() error {
	r.Time = time.Now()
	c := db.GetRedisClient()
	j, err := json.Marshal(r)
	if err != nil {
		log.Printf("fail to save short link request history with key: %s, error: %v\n", r.key(), err)
		return err
	}

	c.LPush(r.key(), j)
	return nil
}
