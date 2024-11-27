package response

import "time"

// TopicResponse 话题响应
type TopicResponse struct {
	ID           uint64       `json:"id"`
	Title        string       `json:"title"`
	Content      string       `json:"content"`
	Location     *Location    `json:"location,omitempty"`
	User         *UserBrief   `json:"user"`
	Images       []TopicImage `json:"images"`
	Tags         []TagInfo    `json:"tags"`
	LikesCount   uint         `json:"likes_count"`
	ViewsCount   uint         `json:"views_count"`
	SharesCount  uint         `json:"shares_count"`
	ExpiresAt    time.Time    `json:"expires_at"`
	Status       string       `json:"status"`
	CreatedAt    time.Time    `json:"created_at"`
	HasLiked     bool         `json:"has_liked"`
	HasFavorited bool         `json:"has_favorited"`
}

// TopicImage 话题图片
type TopicImage struct {
	ID     uint64 `json:"id"`
	URL    string `json:"url"`
	Width  uint   `json:"width"`
	Height uint   `json:"height"`
	Size   uint   `json:"size"`
}

// TagInfo 标签信息
type TagInfo struct {
	ID       uint64 `json:"id"`
	Name     string `json:"name"`
	UseCount uint   `json:"use_count"`
}

// UserBrief 用户简要信息
type UserBrief struct {
	ID        uint64 `json:"id"`
	Nickname  string `json:"nickname"`
	AvatarURL string `json:"avatar_url"`
}

// TopicDetailResponse 话题详情响应
type TopicDetailResponse struct {
	TopicResponse
	Interactions TopicInteractions `json:"interactions"`
}

// TopicInteractions 话题互动统计
type TopicInteractions struct {
	RecentLikes       []UserBrief `json:"recent_likes"`
	RecentShares      []UserBrief `json:"recent_shares"`
	ParticipantsCount uint        `json:"participants_count"`
}
