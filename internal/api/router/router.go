package router

import (
	"time"

	"github.com/chiliososada/distance-back/internal/api/handler"
	"github.com/chiliososada/distance-back/internal/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// SetupRouter 配置路由
func SetupRouter(h *handler.Handler) *gin.Engine {
	r := gin.New()

	// 使用日志和恢复中间件
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	r.Use(middleware.ErrorHandler()) // 错误处理
	// r.Use(middleware.RequestLogger(config))    // 日志记录
	// r.Use(middleware.CORS())                    // 跨域处理
	// r.Use(middleware.Timeout(10 * time.Second)) // 超时控制

	// API限流
	//r.Use(middleware.RateLimit(100, 10)) // 限制每IP每秒请求数

	// CORS 配置
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	// API 版本组
	v1 := r.Group("/api/v1")

	// 认证相关路由
	auth := v1.Group("/auth")
	{
		auth.POST("/register", middleware.AuthRequired(), h.RegisterUser)
	}

	// 需要认证的路由组
	authenticated := v1.Group("")
	authenticated.Use(middleware.AuthRequired())
	{

		// 用户相关路由
		users := authenticated.Group("/users")
		{
			users.GET("/profile", h.GetProfile)      // 获取个人资料
			users.PUT("/profile", h.UpdateProfile)   // 更新个人资料
			users.PUT("/avatar", h.UpdateAvatar)     // 更新头像
			users.PUT("/location", h.UpdateLocation) // 更新位置
			users.GET("/nearby", h.GetNearbyUsers)   // 获取附近用户
			users.POST("/devices", h.RegisterDevice) // 注册设备

			// 用户查询
			users.GET("/search", h.SearchUsers) // 搜索用户
			users.GET("/:id", h.GetUserProfile) // 获取用户资料
		}

		// 关系相关路由
		relationship := authenticated.Group("/relationships")
		{

			// 查询关系

			relationship.GET("/followers", h.GetFollowers)   // 获取粉丝列表
			relationship.GET("/followings", h.GetFollowings) // 获取关注列表
			relationship.GET("/friends", h.GetFriends)       // 获取好友列表

			// 关注相关
			relationship.GET("/:id", h.CheckRelationship)    // 检查与用户的关系状态
			relationship.POST("/:id/follow", h.Follow)       // 关注用户
			relationship.POST("/:id/unfollow", h.Unfollow)   // 取消关注
			relationship.POST("/:id/accept", h.AcceptFollow) // 接受关注请求
			relationship.POST("/:id/reject", h.RejectFollow) // 拒绝关注请求

		}

		// 话题相关路由
		topics := authenticated.Group("/topics")
		{
			// 基础操作
			topics.POST("", h.CreateTopic)       // 创建话题
			topics.PUT("/:id", h.UpdateTopic)    // 更新话题
			topics.DELETE("/:id", h.DeleteTopic) // 删除话题
			topics.GET("/:id", h.GetTopic)       // 获取话题详情

			// 列表查询
			topics.GET("", h.ListTopics)               // 获取话题列表
			topics.GET("/users/:id", h.ListUserTopics) // 获取用户的话题
			topics.GET("/nearby", h.GetNearbyTopics)   // 获取附近话题

			// 图片管理
			topics.POST("/:id/images", h.AddTopicImage) // 添加话题图片

			// 互动相关
			topics.POST("/:id/interactions/:type", h.AddTopicInteraction)      // 添加互动
			topics.DELETE("/:id/interactions/:type", h.RemoveTopicInteraction) // 移除互动
			topics.GET("/:id/interactions/:type", h.GetTopicInteractions)      // 获取互动列表

			// 标签相关路由
			topics.GET("/:id/tags", h.GetTopicTags)
			topics.POST("/:id/tags", h.AddTags)
			topics.DELETE("/:id/tags", h.RemoveTags)
		}

		// 聊天相关路由
		chats := authenticated.Group("/chats")
		{
			// 聊天室管理
			chats.POST("/private/:target_id", h.CreatePrivateRoom) // 创建私聊
			chats.POST("/groups", h.CreateGroupRoom)               // 创建群聊
			chats.GET("", h.ListRooms)                             // 获取聊天室列表
			chats.GET("/:id", h.GetRoomInfo)                       // 获取聊天室信息

			chats.PUT("/:id", h.UpdateRoom)              // 更新聊天室信息
			chats.PUT("/:id/avatar", h.UpdateRoomAvatar) // 更新聊天室头像
			chats.POST("/:id/leave", h.LeaveRoom)        // 离开聊天室
			// 成员管理
			chats.POST("/:id/members", h.AddMember)                       // 添加成员
			chats.DELETE("/:id/members/:member_id", h.RemoveMember)       // 移除成员
			chats.PUT("/:id/members/:member_id/role", h.UpdateMemberRole) // 更新成员角色
			chats.GET("/:id/members", h.GetRoomMembers)                   // 获取聊天室成员列表
			chats.PUT("/:id/members/:member_id", h.UpdateMember)          // 更新成员信息
			// 消息管理
			chats.POST("/:id/messages", h.SendMessage)                         // 发送消息
			chats.GET("/:id/messages", h.GetMessages)                          // 获取消息历史
			chats.POST("/:id/messages/:message_id/read", h.MarkMessagesAsRead) // 标记消息已读
			chats.GET("/:id/unread", h.GetUnreadCount)                         // 获取未读数

			// 其他功能
			chats.POST("/:id/pin", h.PinRoom)     // 置顶聊天室
			chats.DELETE("/:id/pin", h.UnpinRoom) // 取消置顶
		}

		// 添加独立的标签路由组
		tags := authenticated.Group("/tags")
		{
			tags.GET("/popular", h.GetPopularTags)
		}
	}

	return r
}
