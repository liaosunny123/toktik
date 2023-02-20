package model

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"toktik/constant/config"
	"toktik/logging"
)

// User 用户表 /*
type User struct {
	Model                   // 基础模型
	Username        string  `gorm:"not null;unique;size: 32"` // 用户名
	Password        *string `gorm:"not null;size: 32"`        // 密码
	FollowCount     uint32  `gorm:"default:0"`                // 关注总数
	FollowerCount   uint32  `gorm:"default:0"`                // 粉丝总数
	Avatar          *string // 用户头像
	BackgroundImage *string // 背景图片
	Signature       *string // 个人简介
	TotalFavorited  *uint32 `gorm:"default:0"`       // 获赞数量
	WorkCount       *uint32 `gorm:"default:0"`       // 作品数量
	FavoriteCount   *uint32 `gorm:"default:0"`       // 点赞数量
	Name            string  `gorm:"not null;unique"` // 用户名称
	Role            string  `gorm:"default:0"`       // 用户角色

	updated bool
}

func (u *User) IsUpdated() bool {
	return u.updated
}

func (u *User) isEmail() bool {
	parts := strings.Split(u.Username, "@")
	if len(parts) != 2 {
		return false
	}

	userPart := parts[0]
	if len(userPart) == 0 {
		return false
	}

	domainPart := parts[1]
	if len(domainPart) < 3 {
		return false
	}
	if !strings.Contains(domainPart, ".") {
		return false
	}
	if strings.HasPrefix(domainPart, ".") || strings.HasSuffix(domainPart, ".") {
		return false
	}

	return true
}

func getEmailMD5(email string) (md5String string) {
	lowerEmail := strings.ToLower(email)
	hasher := md5.New()
	hasher.Write([]byte(lowerEmail))
	md5Bytes := hasher.Sum(nil)
	md5String = hex.EncodeToString(md5Bytes)
	return
}

type unsplashResponse struct {
	Urls struct {
		Regular string `json:"regular"`
	} `json:"urls"`
}

func getImageFromUnsplash(query string) (url string, err error) {
	unsplashUrl := fmt.Sprintf("https://api.unsplash.com/photos/random?query=%s&count=1", query)

	client := &http.Client{}
	req, _ := http.NewRequest("GET", unsplashUrl, nil)
	req.Header.Add("Authorization", "Client-ID "+config.EnvConfig.UNSPLASH_ACCESS_KEY)
	resp, _ := client.Do(req)

	if resp.Body != nil {
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				logging.Logger.Errorf("getImageFromUnsplash: %v", err)
			}
		}(resp.Body)
	}
	body, _ := io.ReadAll(resp.Body)

	var response unsplashResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return
	}

	url = response.Urls.Regular
	return
}

func getCravatarUrl(email string) string {
	return `https://cravatar.cn/avatar/` + getEmailMD5(email) + `?d=` + "identicon"
}

func (u *User) GetBackgroundImage() (url string) {
	if u.BackgroundImage != nil {
		return *u.BackgroundImage
	}

	defer func() {
		u.BackgroundImage = &url
		u.updated = true
	}()

	unsplashURL, err := getImageFromUnsplash(u.Username)
	if err != nil {
		catURL, err := getImageFromUnsplash("cat")
		if err != nil {
			return getCravatarUrl(u.Username)
		}
		return catURL
	}
	return unsplashURL
}

func (u *User) GetUserAvatar() (url string) {
	if u.Avatar != nil {
		return *u.Avatar
	}

	defer func() {
		u.Avatar = &url
		u.updated = true
	}()

	if u.isEmail() {
		return getCravatarUrl(u.Username)
	}

	unsplashURL, err := getImageFromUnsplash(u.Username)
	if err != nil || unsplashURL == "" {
		catURL, err := getImageFromUnsplash("cat")
		if err != nil || catURL == "" {
			return getCravatarUrl(u.Username)
		}
		return catURL
	}

	return unsplashURL
}
