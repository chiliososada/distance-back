GRANT ALL ON *.* TO root@'%';

-- 用户基础信息表
CREATE TABLE users (
    id BIGINT UNSIGNED PRIMARY KEY COMMENT '用户ID',
    nickname VARCHAR(50) COMMENT '用户昵称',
    avatar_url VARCHAR(255) COMMENT '头像URL',
    birth_date DATE COMMENT '出生日期',
    gender ENUM('male', 'female', 'other') COMMENT '性别',
    bio TEXT COMMENT '个人简介',
    location_latitude DECIMAL(10, 8) COMMENT '位置纬度 profile',
    location_longitude DECIMAL(11, 8) COMMENT '位置经度 profile',
    language VARCHAR(10) DEFAULT 'zh_CN' COMMENT '语言设置',
    status ENUM('active', 'inactive', 'banned') DEFAULT 'active' COMMENT '账号状态',
    privacy_level ENUM('public', 'friends', 'private') DEFAULT 'public' COMMENT '隐私级别 好友可见 全部都可以看见 只有自己能看', 
    notification_enabled BOOLEAN DEFAULT TRUE COMMENT '是否开启通知',
    location_sharing BOOLEAN DEFAULT TRUE COMMENT '是否开启位置共享',
    photo_enabled BOOLEAN DEFAULT TRUE COMMENT '是否开启图片访问',
    last_active_at TIMESTAMP COMMENT '最后活跃时间',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX idx_location (location_latitude, location_longitude),
    user_type ENUM('individual', 'merchant', 'official', 'admin') DEFAULT 'individual' COMMENT '用户类型：individual-普通用户, merchant-商家, official-官方账号,admin-管理员',
    INDEX idx_status (status)
) COMMENT '用户基础信息表';

-- 管理员权限表
CREATE TABLE admin_permissions (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT UNSIGNED NOT NULL COMMENT '管理员用户ID',
    permission_type ENUM(
    	'super_admin',      -- 超级管理员(最高权限)
        'user_manage',      -- 用户管理
        'content_audit',    -- 内容审核
        'system_config',    -- 系统配置
        'data_analysis'     -- 数据分析
    ) NOT NULL COMMENT '权限类型',
    status BOOLEAN DEFAULT TRUE COMMENT '是否启用',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id),
    UNIQUE KEY uk_admin_permission (user_id, permission_type)
) COMMENT '管理员权限表';

---被办用户
CREATE TABLE user_bans (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT UNSIGNED NOT NULL COMMENT '被封禁用户ID',
    operator_id BIGINT UNSIGNED NOT NULL COMMENT '操作人ID', 
    reason TEXT NOT NULL COMMENT '封禁原因',
    ban_start TIMESTAMP NOT NULL COMMENT '封禁开始时间',
    ban_end TIMESTAMP COMMENT '封禁结束时间',
    status ENUM('active', 'expired', 'cancelled') DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (operator_id) REFERENCES users(id),
    INDEX idx_user_status (user_id, status)
) COMMENT '用户封禁记录表';
-- 用户认证信息表（Firebase）
CREATE TABLE user_authentications (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT COMMENT '认证ID',
    user_id BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    firebase_uid VARCHAR(128) NOT NULL COMMENT 'Firebase用户唯一标识',
    auth_provider ENUM('password', 'phone', 'google', 'apple', 'anonymous') NOT NULL COMMENT '认证方式',
    email VARCHAR(100) COMMENT '邮箱地址',
    phone_number VARCHAR(20) COMMENT '手机号码',
    last_sign_in_at TIMESTAMP COMMENT '最后登录时间',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE KEY uk_firebase_uid (firebase_uid),
    UNIQUE KEY uk_email (email),
    UNIQUE KEY uk_phone (phone_number),
    INDEX idx_auth_provider (auth_provider)
) COMMENT '用户认证信息表';

-- 用户设备表（推送相关）--待定 一个用户多个设备
CREATE TABLE user_devices (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT COMMENT '设备ID',
    user_id BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    
    -- 推送相关
    device_token VARCHAR(255) NOT NULL COMMENT '设备推送令牌',
    push_provider ENUM('fcm', 'apns', 'web') NOT NULL COMMENT '推送提供商 apns ios fcm 安卓 web websock',
    push_enabled BOOLEAN DEFAULT TRUE COMMENT '是否允许推送',
    
    -- 设备信息
    device_type ENUM('ios', 'android', 'web') NOT NULL COMMENT '设备类型',
    device_name VARCHAR(100) COMMENT '设备名称',
    device_model VARCHAR(50) COMMENT '设备型号',
    os_version VARCHAR(20) COMMENT '系统版本',
    app_version VARCHAR(20) COMMENT 'APP版本号',
    browser_info VARCHAR(200) COMMENT '浏览器信息(web端)',
    browser_fingerprint VARCHAR(100) COMMENT '浏览器指纹(web端)',
    
    -- 状态信息
    is_active BOOLEAN DEFAULT TRUE COMMENT '是否活跃',
    last_active_at TIMESTAMP COMMENT '最后活跃时间',
    badge_count INT UNSIGNED DEFAULT 0 COMMENT '未读消息数量',
    
    -- 时间戳
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    
    -- 关联和索引
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE KEY uk_device_token (device_token),
    INDEX idx_user_device (user_id, device_type, is_active),
    INDEX idx_browser_fingerprint (browser_fingerprint)
) COMMENT '用户设备推送表';


CREATE TABLE user_relationships (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT COMMENT '关系ID',
    follower_id BIGINT UNSIGNED NOT NULL COMMENT '关注人/请求人用户ID',
    following_id BIGINT UNSIGNED NOT NULL COMMENT '被关注人/被请求人用户ID',
    status ENUM('pending', 'accepted', 'blocked') DEFAULT 'pending' COMMENT '关系状态：pending-等待确认, accepted-已确认, blocked-已屏蔽',

    -- 时间相关
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '关系创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '关系更新时间',
    accepted_at TIMESTAMP NULL DEFAULT NULL COMMENT '关系确认时间',

    -- 外键约束
    FOREIGN KEY (follower_id) REFERENCES users(id) ON DELETE CASCADE COMMENT '关联关注人用户表',
    FOREIGN KEY (following_id) REFERENCES users(id) ON DELETE CASCADE COMMENT '关联被关注人用户表',

    -- 唯一约束
    UNIQUE KEY unique_relationship (follower_id, following_id) COMMENT '保证两个用户之间只有一条关系记录',

    -- 索引
    INDEX idx_follower (follower_id, status) COMMENT '查询用户关注列表',
    INDEX idx_following (following_id, status) COMMENT '查询用户粉丝列表'
) COMMENT '用户关系表';

-- 话题表
CREATE TABLE topics (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT COMMENT '话题ID',
    user_id BIGINT UNSIGNED NOT NULL COMMENT '发布者用户ID',
    title VARCHAR(255) NOT NULL COMMENT '话题标题',
    content TEXT COMMENT '话题内容',
    location_latitude DECIMAL(10, 8) COMMENT '位置纬度',
    location_longitude DECIMAL(11, 8) COMMENT '位置经度',
    likes_count INT UNSIGNED DEFAULT 0 COMMENT '点赞数',
    participants_count INT UNSIGNED DEFAULT 0 COMMENT '参与人数',
    views_count INT UNSIGNED DEFAULT 0 COMMENT '浏览次数',
    shares_count INT UNSIGNED DEFAULT 0 COMMENT '分享次数',
    expires_at TIMESTAMP COMMENT '过期时间',
    status ENUM('active', 'closed', 'cancelled') DEFAULT 'active'  COMMENT '话题状态：active-进行中, closed-已结束, cancelled-已取消',
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    -- 外键约束
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE COMMENT '关联用户表',
    -- 索引
    INDEX idx_location (location_latitude, location_longitude) COMMENT '位置索引',
    INDEX idx_status (status) COMMENT '状态索引',
    INDEX idx_user_time (user_id, created_at) COMMENT '用户话题时间索引'
) COMMENT '话题表';

-- 话题图片表
CREATE TABLE topic_images (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT COMMENT '图片ID',
    topic_id BIGINT UNSIGNED NOT NULL COMMENT '话题ID',
    image_url VARCHAR(255) NOT NULL COMMENT '图片URL',
    sort_order INT UNSIGNED DEFAULT 0 COMMENT '排序顺序',
    image_width INT UNSIGNED COMMENT '图片宽度',
    image_height INT UNSIGNED COMMENT '图片高度',
    file_size INT UNSIGNED COMMENT '文件大小(字节)',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    
    -- 外键约束
    FOREIGN KEY (topic_id) REFERENCES topics(id) ON DELETE CASCADE COMMENT '关联话题表',
    
    -- 索引
    INDEX idx_topic_sort (topic_id, sort_order) COMMENT '话题图片排序索引'
) COMMENT '话题图片表';

-- 话题标签表
CREATE TABLE tags (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT COMMENT '标签ID',
    name VARCHAR(50) UNIQUE NOT NULL COMMENT '标签名称',
    use_count INT UNSIGNED DEFAULT 0 COMMENT '使用次数',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    
    -- 索引
    INDEX idx_use_count (use_count) COMMENT '使用次数索引，用于热门标签查询'
) COMMENT '话题标签表';

-- 话题-标签关联表
CREATE TABLE topic_tags (
    topic_id BIGINT UNSIGNED COMMENT '话题ID',
    tag_id BIGINT UNSIGNED COMMENT '标签ID',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    
    -- 主键
    PRIMARY KEY (topic_id, tag_id) COMMENT '联合主键，确保话题和标签的唯一关联',
    
    -- 外键约束
    FOREIGN KEY (topic_id) REFERENCES topics(id) ON DELETE CASCADE COMMENT '关联话题表',
    FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE COMMENT '关联标签表',
    
    -- 索引
    INDEX idx_tag_time (tag_id, created_at) COMMENT '标签时间索引，用于查询标签下的最新话题'
) COMMENT '话题-标签关联表';

-- 话题互动表
CREATE TABLE topic_interactions (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT COMMENT '互动ID',
    topic_id BIGINT UNSIGNED NOT NULL COMMENT '话题ID',
    user_id BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    interaction_type ENUM('like', 'favorite', 'share') NOT NULL 
        COMMENT '互动类型：like-点赞, favorite-收藏, share-分享',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '互动时间',
    interaction_status ENUM('active', 'cancelled') DEFAULT 'active' COMMENT '互动状态：active-有效, cancelled-已撤销',
    -- 外键约束
    FOREIGN KEY (topic_id) REFERENCES topics(id) ON DELETE CASCADE COMMENT '关联话题表',
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE COMMENT '关联用户表',
    
    -- 唯一约束
    UNIQUE KEY unique_interaction (topic_id, user_id, interaction_type) COMMENT '确保用户对同一话题的同类互动只有一次',
    
    -- 索引
    INDEX idx_user_type (user_id, interaction_type, created_at) COMMENT '用户互动类型索引，用于查询用户的互动历史'
) COMMENT '话题互动表';


---聊天
-- 聊天室表
CREATE TABLE chat_rooms (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT COMMENT '聊天室ID',
    name VARCHAR(100) NOT NULL COMMENT '聊天室名称',
    type ENUM('individual', 'group', 'merchant', 'official') NOT NULL COMMENT '聊天室类型：individual-个人聊天, group-群聊, merchant-店家, official-官方',
    topic_id BIGINT UNSIGNED COMMENT '关联的话题ID（如果是话题聊天室）',
    avatar_url VARCHAR(255) COMMENT '聊天室头像URL',
    announcement TEXT COMMENT '聊天室公告',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '聊天室创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '聊天室更新时间',
    FOREIGN KEY (topic_id) REFERENCES topics(id) COMMENT '关联话题表'
) COMMENT '聊天室表，用于记录聊天室的基本信息';

-- 聊天室成员表
CREATE TABLE chat_room_members (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT COMMENT '成员记录ID',
    chat_room_id BIGINT UNSIGNED COMMENT '聊天室ID',
    user_id BIGINT UNSIGNED COMMENT '用户ID',
    role ENUM('owner', 'admin', 'member') DEFAULT 'member' COMMENT '成员角色：owner-拥有者, admin-管理员, member-普通成员',
    nickname VARCHAR(50) COMMENT '成员在聊天室中的昵称',
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '加入聊天室时间',
    last_read_message_id BIGINT UNSIGNED DEFAULT 0 COMMENT '最后读取的消息ID',
    is_muted BOOLEAN DEFAULT FALSE COMMENT '是否被静音',
    FOREIGN KEY (chat_room_id) REFERENCES chat_rooms(id) COMMENT '关联聊天室表',
    FOREIGN KEY (user_id) REFERENCES users(id) COMMENT '关联用户表',
    UNIQUE KEY unique_member (chat_room_id, user_id) COMMENT '唯一约束：同一聊天室中的用户不能重复'
) COMMENT '聊天室成员表，用于记录聊天室的成员信息';

-- 消息表
CREATE TABLE messages (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT COMMENT '消息ID',
    chat_room_id BIGINT UNSIGNED COMMENT '所属聊天室ID',
    sender_id BIGINT UNSIGNED COMMENT '发送者用户ID',
   content_type ENUM('text', 'image', 'file', 'system') DEFAULT 'text' NOT NULL COMMENT '消息类型：text-文本, image-图片, file-文件, system-系统消息',
    content TEXT COMMENT '消息内容',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '消息发送时间',
    FOREIGN KEY (chat_room_id) REFERENCES chat_rooms(id) COMMENT '关联聊天室表',
    FOREIGN KEY (sender_id) REFERENCES users(id) COMMENT '关联用户表',
    INDEX idx_chat_room_time (chat_room_id, created_at) COMMENT '聊天室和时间索引，用于按时间查询消息'
) COMMENT '消息表，用于记录聊天室中的消息内容';

-- 消息媒体表
CREATE TABLE message_media (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT COMMENT '媒体记录ID',
    message_id BIGINT UNSIGNED COMMENT '关联的消息ID',
    media_type ENUM('image', 'file') NOT NULL COMMENT '媒体类型：image-图片, file-文件',
    media_url VARCHAR(255) NOT NULL COMMENT '媒体文件的URL',
    file_name VARCHAR(255) COMMENT '媒体文件名称',
    file_size INT UNSIGNED COMMENT '媒体文件大小（单位：字节）',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '媒体文件上传时间',
    FOREIGN KEY (message_id) REFERENCES messages(id) COMMENT '关联消息表'
) COMMENT '消息媒体表，用于存储消息中包含的媒体文件信息';

-- 聊天室置顶表
CREATE TABLE pinned_chat_rooms (
    user_id BIGINT UNSIGNED COMMENT '用户ID',
    chat_room_id BIGINT UNSIGNED COMMENT '聊天室ID',
    pinned_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '置顶时间',
    PRIMARY KEY (user_id, chat_room_id) COMMENT '主键：用户ID和聊天室ID联合唯一',
    FOREIGN KEY (user_id) REFERENCES users(id) COMMENT '关联用户表',
    FOREIGN KEY (chat_room_id) REFERENCES chat_rooms(id) COMMENT '关联聊天室表'
) COMMENT '聊天室置顶表，用于记录用户置顶的聊天室';

----
-- 通知模板表
CREATE TABLE notification_templates (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT COMMENT '模板ID',
    code VARCHAR(50) UNIQUE NOT NULL COMMENT '模板代码',
    title_template TEXT NOT NULL COMMENT '标题模板',
    content_template TEXT NOT NULL COMMENT '内容模板',
    platform ENUM('all', 'ios', 'android', 'web') DEFAULT 'all' COMMENT '平台类型',
    variables JSON COMMENT '模板变量',
    status ENUM('active', 'inactive') DEFAULT 'active' COMMENT '模板状态',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间'
) COMMENT '通知模板表';

-- 推送消息表
CREATE TABLE push_notifications (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT COMMENT '推送ID',
    template_id BIGINT UNSIGNED COMMENT '模板ID',
    title VARCHAR(255) NOT NULL COMMENT '推送标题',
    content TEXT NOT NULL COMMENT '推送内容',
    type ENUM('system', 'merchant', 'topic', 'chat') NOT NULL COMMENT '推送类型',
    sender_id BIGINT UNSIGNED COMMENT '发送者ID',
    target_type ENUM('all', 'area', 'specific_users') NOT NULL COMMENT '目标类型',
    target_params JSON COMMENT '目标参数',
    status ENUM('draft', 'scheduled', 'sending', 'sent', 'cancelled') COMMENT '推送状态',
    scheduled_at TIMESTAMP COMMENT '计划发送时间',
    sent_at TIMESTAMP COMMENT '实际发送时间',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    FOREIGN KEY (template_id) REFERENCES notification_templates(id),
    FOREIGN KEY (sender_id) REFERENCES users(id)
) COMMENT '推送消息表';

-- 推送接收记录表
CREATE TABLE push_notification_recipients (
    notification_id BIGINT UNSIGNED COMMENT '推送ID',
    user_id BIGINT UNSIGNED COMMENT '接收者ID',
    device_id BIGINT UNSIGNED COMMENT '接收设备ID',
    status ENUM('pending', 'sent', 'failed', 'received', 'read') COMMENT '接收状态',
    error_message TEXT COMMENT '错误信息',
    sent_at TIMESTAMP COMMENT '发送时间',
    received_at TIMESTAMP COMMENT '接收时间',
    read_at TIMESTAMP COMMENT '阅读时间',
    PRIMARY KEY (notification_id, user_id),
    FOREIGN KEY (notification_id) REFERENCES push_notifications(id),
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (device_id) REFERENCES user_devices(id)
) COMMENT '推送接收记录表';


-- 地区表 --待定
CREATE TABLE regions (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT COMMENT '地区ID',
    parent_id BIGINT UNSIGNED COMMENT '父级地区ID',
    name VARCHAR(50) NOT NULL COMMENT '地区名称',
    level ENUM('country', 'province', 'city', 'district') NOT NULL COMMENT '地区级别',
    code VARCHAR(20) UNIQUE COMMENT '地区编码',
    latitude DECIMAL(10, 8) COMMENT '中心纬度',
    longitude DECIMAL(11, 8) COMMENT '中心经度',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    FOREIGN KEY (parent_id) REFERENCES regions(id)
) COMMENT '地区表';

-- 用户位置历史表
CREATE TABLE user_locations (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT COMMENT '记录ID',
    user_id BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    latitude DECIMAL(10, 8) NOT NULL COMMENT '纬度',
    longitude DECIMAL(11, 8) NOT NULL COMMENT '经度',
    region_id BIGINT UNSIGNED COMMENT '所属地区ID',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '记录时间',
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (region_id) REFERENCES regions(id),
    INDEX idx_user_location (user_id, created_at)
) COMMENT '用户位置历史表';

-- 敏感词表 --前期不要
CREATE TABLE sensitive_words (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT COMMENT '敏感词ID',
    word VARCHAR(50) NOT NULL UNIQUE COMMENT '敏感词',
    category VARCHAR(20) COMMENT '分类',
    level ENUM('low', 'medium', 'high') DEFAULT 'medium' COMMENT '级别',
    status BOOLEAN DEFAULT TRUE COMMENT '是否启用',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间'
) COMMENT '敏感词表';

-- 举报记录表 --前期不要
CREATE TABLE reports (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT COMMENT '举报ID',
    reporter_id BIGINT UNSIGNED NOT NULL COMMENT '举报人ID',
    target_type ENUM('user', 'topic', 'comment', 'message') NOT NULL COMMENT '举报对象类型',
    target_id BIGINT UNSIGNED NOT NULL COMMENT '举报对象ID',
    reason_type VARCHAR(50) NOT NULL COMMENT '举报原因类型',
    reason_detail TEXT COMMENT '举报详细原因',
    status ENUM('pending', 'processing', 'resolved', 'rejected') DEFAULT 'pending' COMMENT '处理状态',
    handler_id BIGINT UNSIGNED COMMENT '处理人ID',
    handling_result TEXT COMMENT '处理结果',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '举报时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    FOREIGN KEY (reporter_id) REFERENCES users(id),
    FOREIGN KEY (handler_id) REFERENCES users(id),
    INDEX idx_target (target_type, target_id),
    INDEX idx_status (status)
) COMMENT '举报记录表';

-- 为经常进行范围查询的字段建立符合索引
CREATE INDEX idx_topic_location_time ON topics(location_latitude, location_longitude, created_at);
CREATE INDEX idx_user_active_location ON users(status, location_latitude, location_longitude);

-- 为全文搜索字段建立全文索引
ALTER TABLE topics ADD FULLTEXT INDEX ft_topic_content(title, content);

-- 为经常统计的字段建立统计索引
CREATE INDEX idx_topic_stats ON topics(status, created_at);
CREATE INDEX idx_message_stats ON messages(status, created_at);
