

-- Create database with proper charset
CREATE DATABASE IF NOT EXISTS distance_back
DEFAULT CHARACTER SET utf8mb4
DEFAULT COLLATE utf8mb4_unicode_ci;

-- Create user and grant privileges
CREATE USER IF NOT EXISTS 'distance_user'@'%' IDENTIFIED BY 'distance_password';
GRANT ALL PRIVILEGES ON distance_back.* TO 'distance_user'@'%';
FLUSH PRIVILEGES;

USE distance_back;

-- Users table
CREATE TABLE users (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT COMMENT '用户ID',
    uid VARCHAR(36) NOT NULL DEFAULT (UUID()) COMMENT '用户UUID',
    nickname VARCHAR(50) COMMENT '用户昵称',
    avatar_url VARCHAR(255) COMMENT '头像URL',
    birth_date DATE COMMENT '出生日期',
    gender ENUM('male', 'female', 'other') COMMENT '性别',
    bio TEXT COMMENT '个人简介',
    location_latitude DECIMAL(10, 8) COMMENT '位置纬度',
    location_longitude DECIMAL(11, 8) COMMENT '位置经度',
    language VARCHAR(10) DEFAULT 'zh_CN' COMMENT '语言设置',
    status ENUM('active', 'inactive', 'banned') DEFAULT 'active' COMMENT '账号状态',
    privacy_level ENUM('public', 'friends', 'private') DEFAULT 'public' COMMENT '隐私级别',
    notification_enabled BOOLEAN DEFAULT TRUE COMMENT '是否开启通知',
    location_sharing BOOLEAN DEFAULT TRUE COMMENT '是否开启位置共享',
    photo_enabled BOOLEAN DEFAULT TRUE COMMENT '是否开启图片访问',
    last_active_at TIMESTAMP COMMENT '最后活跃时间',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    user_type ENUM('individual', 'merchant', 'official', 'admin') DEFAULT 'individual' COMMENT '用户类型',
    UNIQUE KEY uk_uid (uid),
    INDEX idx_location (location_latitude, location_longitude),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT '用户基础信息表';

-- Admin permissions table
CREATE TABLE admin_permissions (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT COMMENT '权限ID',
    uid VARCHAR(36) NOT NULL DEFAULT (UUID()) COMMENT '权限UUID',
    user_uid VARCHAR(36) NOT NULL COMMENT '用户UUID',
    permission_type ENUM(
        'super_admin',
        'user_manage',
        'content_audit',
        'system_config',
        'data_analysis'
    ) NOT NULL COMMENT '权限类型',
    status BOOLEAN DEFAULT TRUE COMMENT '是否启用',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    UNIQUE KEY uk_uid (uid),
    FOREIGN KEY (user_uid) REFERENCES users(uid),
    UNIQUE KEY uk_admin_permission (user_uid, permission_type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT '管理员权限表';

-- User bans table
CREATE TABLE user_bans (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT COMMENT '封禁记录ID',
    uid VARCHAR(36) NOT NULL DEFAULT (UUID()) COMMENT '封禁UUID',
    user_uid VARCHAR(36) NOT NULL COMMENT '被封禁用户UUID',
    operator_uid VARCHAR(36) NOT NULL COMMENT '操作人UUID',
    reason TEXT NOT NULL COMMENT '封禁原因',
    ban_start TIMESTAMP NOT NULL COMMENT '封禁开始时间',
    ban_end TIMESTAMP COMMENT '封禁结束时间',
    status ENUM('active', 'expired', 'cancelled') DEFAULT 'active' COMMENT '状态',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    UNIQUE KEY uk_uid (uid),
    FOREIGN KEY (user_uid) REFERENCES users(uid),
    FOREIGN KEY (operator_uid) REFERENCES users(uid),
    INDEX idx_user_status (user_uid, status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT '用户封禁记录表';

-- User authentications table
CREATE TABLE user_authentications (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT COMMENT '认证ID',
    uid VARCHAR(36) NOT NULL DEFAULT (UUID()) COMMENT '认证UUID',
    user_uid VARCHAR(36) NOT NULL COMMENT '用户UUID',
    firebase_uid VARCHAR(128) NOT NULL COMMENT 'Firebase用户唯一标识',
    auth_provider ENUM('password', 'phone', 'google', 'apple', 'anonymous') NOT NULL COMMENT '认证方式',
    email VARCHAR(100) COMMENT '邮箱地址',
    phone_number VARCHAR(20) COMMENT '手机号码',
    last_sign_in_at TIMESTAMP COMMENT '最后登录时间',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    UNIQUE KEY uk_uid (uid),
    FOREIGN KEY (user_uid) REFERENCES users(uid) ON DELETE CASCADE,
    UNIQUE KEY uk_firebase_uid (firebase_uid),
    UNIQUE KEY uk_email (email),
    UNIQUE KEY uk_phone (phone_number),
    INDEX idx_auth_provider (auth_provider)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT '用户认证信息表';

-- User devices table
CREATE TABLE user_devices (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT COMMENT '设备ID',
    uid VARCHAR(36) NOT NULL DEFAULT (UUID()) COMMENT '设备UUID',
    user_uid VARCHAR(36) NOT NULL COMMENT '用户UUID',
    device_token VARCHAR(255) NOT NULL COMMENT '设备推送令牌',
    push_provider ENUM('fcm', 'apns', 'web') NOT NULL COMMENT '推送提供商',
    push_enabled BOOLEAN DEFAULT TRUE COMMENT '是否允许推送',
    device_type ENUM('ios', 'android', 'web') NOT NULL COMMENT '设备类型',
    device_name VARCHAR(100) COMMENT '设备名称',
    device_model VARCHAR(50) COMMENT '设备型号',
    os_version VARCHAR(20) COMMENT '系统版本',
    app_version VARCHAR(20) COMMENT 'APP版本号',
    browser_info VARCHAR(200) COMMENT '浏览器信息',
    browser_fingerprint VARCHAR(100) COMMENT '浏览器指纹',
    is_active BOOLEAN DEFAULT TRUE COMMENT '是否活跃',
    last_active_at TIMESTAMP COMMENT '最后活跃时间',
    badge_count INT UNSIGNED DEFAULT 0 COMMENT '未读消息数量',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    UNIQUE KEY uk_uid (uid),
    FOREIGN KEY (user_uid) REFERENCES users(uid) ON DELETE CASCADE,
    UNIQUE KEY uk_device_token (device_token),
    INDEX idx_user_device (user_uid, device_type, is_active),
    INDEX idx_browser_fingerprint (browser_fingerprint)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT '用户设备推送表';

-- User relationships table
CREATE TABLE user_relationships (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT COMMENT '关系ID',
    uid VARCHAR(36) NOT NULL DEFAULT (UUID()) COMMENT '关系UUID',
    follower_uid VARCHAR(36) NOT NULL COMMENT '关注人UUID',
    following_uid VARCHAR(36) NOT NULL COMMENT '被关注人UUID',
    status ENUM('pending', 'accepted', 'rejected', 'blocked') DEFAULT 'pending' COMMENT '关系状态：pending-等待确认, accepted-已确认, rejected-已拒绝, rejected-已拉黑',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '关系创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '关系更新时间',
    accepted_at TIMESTAMP NULL DEFAULT NULL COMMENT '关系确认时间',
    UNIQUE KEY uk_uid (uid),
    FOREIGN KEY (follower_uid) REFERENCES users(uid),
    FOREIGN KEY (following_uid) REFERENCES users(uid),
    UNIQUE KEY unique_relationship (follower_uid, following_uid),
    INDEX idx_follower (follower_uid, status),
    INDEX idx_following (following_uid, status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT '用户关系表';

-- Topics table
CREATE TABLE topics (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT COMMENT '话题ID',
    uid VARCHAR(36) NOT NULL DEFAULT (UUID()) COMMENT '话题UUID',
    user_uid VARCHAR(36) NOT NULL COMMENT '发布者UUID',
    title VARCHAR(255) NOT NULL COMMENT '话题标题',
    content TEXT COMMENT '话题内容',
    location_latitude DECIMAL(10, 8) COMMENT '位置纬度',
    location_longitude DECIMAL(11, 8) COMMENT '位置经度',
    likes_count INT UNSIGNED DEFAULT 0 COMMENT '点赞数',
    participants_count INT UNSIGNED DEFAULT 0 COMMENT '参与人数',
    views_count INT UNSIGNED DEFAULT 0 COMMENT '浏览次数',
    shares_count INT UNSIGNED DEFAULT 0 COMMENT '分享次数',
    expires_at TIMESTAMP COMMENT '过期时间',
    status ENUM('active', 'closed', 'cancelled') DEFAULT 'active' COMMENT '话题状态',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    UNIQUE KEY uk_uid (uid),
    FOREIGN KEY (user_uid) REFERENCES users(uid),
    INDEX idx_location (location_latitude, location_longitude),
    INDEX idx_status (status),
    INDEX idx_user_time (user_uid, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT '话题表';

-- Topic images table
CREATE TABLE topic_images (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT COMMENT '图片ID',
    uid VARCHAR(36) NOT NULL DEFAULT (UUID()) COMMENT '图片UUID',
    topic_uid VARCHAR(36) NOT NULL COMMENT '话题UUID',
    image_url VARCHAR(255) NOT NULL COMMENT '图片URL',
    sort_order INT UNSIGNED DEFAULT 0 COMMENT '排序顺序',
    image_width INT UNSIGNED COMMENT '图片宽度',
    image_height INT UNSIGNED COMMENT '图片高度',
    file_size INT UNSIGNED COMMENT '文件大小(字节)',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    UNIQUE KEY uk_uid (uid),
    FOREIGN KEY (topic_uid) REFERENCES topics(uid),
    INDEX idx_topic_sort (topic_uid, sort_order)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT '话题图片表';

-- Tags table
CREATE TABLE tags (
    uid VARCHAR(36) NOT NULL DEFAULT (UUID()) COMMENT '标签UUID',
    name VARCHAR(50) UNIQUE NOT NULL COMMENT '标签名称',
    use_count INT UNSIGNED DEFAULT 0 COMMENT '使用次数',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    UNIQUE KEY uk_uid (uid),
    INDEX idx_use_count (use_count)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT '标签表';

-- Topic tags relation table
CREATE TABLE topic_tags (
    topic_uid VARCHAR(36) NOT NULL COMMENT '话题UUID',
    tag_uid VARCHAR(36) NOT NULL COMMENT '标签UUID',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (topic_uid, tag_uid),
    FOREIGN KEY (topic_uid) REFERENCES topics(uid),
    FOREIGN KEY (tag_uid) REFERENCES tags(uid),
    INDEX idx_tag_time (tag_uid, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT '话题-标签关联表';

-- Topic interactions table
CREATE TABLE topic_interactions (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT COMMENT '互动ID',
    uid VARCHAR(36) NOT NULL DEFAULT (UUID()) COMMENT '互动UUID',
    topic_uid VARCHAR(36) NOT NULL COMMENT '话题UUID',
    user_uid VARCHAR(36) NOT NULL COMMENT '用户UUID',
    interaction_type ENUM('like', 'favorite', 'share') NOT NULL COMMENT '互动类型',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '互动时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    interaction_status ENUM('active', 'cancelled') DEFAULT 'active' COMMENT '互动状态',
    UNIQUE KEY uk_uid (uid),
    FOREIGN KEY (topic_uid) REFERENCES topics(uid),
    FOREIGN KEY (user_uid) REFERENCES users(uid),
    UNIQUE KEY unique_interaction (topic_uid, user_uid, interaction_type),
    INDEX idx_user_type (user_uid, interaction_type, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT '话题互动表';

-- Chat rooms table
CREATE TABLE chat_rooms (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT COMMENT '聊天室ID',
    uid VARCHAR(36) NOT NULL DEFAULT (UUID()) COMMENT '聊天室UUID',
    name VARCHAR(100) NOT NULL COMMENT '聊天室名称',
    type ENUM('individual', 'group', 'merchant', 'official') NOT NULL COMMENT '聊天室类型',
    topic_uid VARCHAR(36) NULL COMMENT '关联话题UUID',    -- 只设置为可以为 NULL
    avatar_url VARCHAR(255) COMMENT '聊天室头像URL',
    announcement TEXT COMMENT '聊天室公告',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    UNIQUE KEY uk_uid (uid)
    -- 去掉 FOREIGN KEY 约束
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT '聊天室表';

CREATE TABLE chat_rooms (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT COMMENT '聊天室ID',
    uid VARCHAR(36) NOT NULL DEFAULT (UUID()) COMMENT '聊天室UUID',
    name VARCHAR(100) NOT NULL COMMENT '聊天室名称',
    type ENUM('individual', 'group', 'merchant', 'official') NOT NULL COMMENT '聊天室类型',
    topic_uid VARCHAR(36) NULL COMMENT '关联话题UUID',
    avatar_url VARCHAR(255) COMMENT '聊天室头像URL',
    announcement TEXT COMMENT '聊天室公告',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    UNIQUE KEY uk_uid (uid),
    FOREIGN KEY (topic_uid) REFERENCES topics(uid) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT '聊天室表';

-- Chat room members table
CREATE TABLE chat_room_members (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT COMMENT '成员记录ID',
    uid VARCHAR(36) NOT NULL DEFAULT (UUID()) COMMENT '成员UUID',
    chat_room_uid VARCHAR(36) NOT NULL COMMENT '聊天室UUID',
    user_uid VARCHAR(36) NOT NULL COMMENT '用户UUID',
    role ENUM('owner', 'admin', 'member') DEFAULT 'member' COMMENT '成员角色',
    nickname VARCHAR(50) COMMENT '成员在聊天室中的昵称',
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '加入时间',
    last_read_message_id VARCHAR(36) NOT NULL DEFAULT '00000000-0000-0000-0000-000000000000' COMMENT '最后读取的消息UUID',
    is_muted BOOLEAN DEFAULT FALSE COMMENT '是否被静音',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    UNIQUE KEY uk_uid (uid),
    FOREIGN KEY (chat_room_uid) REFERENCES chat_rooms(uid),
    FOREIGN KEY (user_uid) REFERENCES users(uid),
    UNIQUE KEY unique_member (chat_room_uid, user_uid)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT '聊天室成员表';

-- Messages table
CREATE TABLE messages (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT COMMENT '消息ID',
    uid VARCHAR(36) NOT NULL DEFAULT (UUID()) COMMENT '消息UUID',
    chat_room_uid VARCHAR(36) NOT NULL COMMENT '聊天室UUID',
    sender_uid VARCHAR(36) NOT NULL COMMENT '发送者UUID',
    content_type ENUM('text', 'image', 'file', 'system') DEFAULT 'text' NOT NULL COMMENT '消息类型',
    content TEXT COMMENT '消息内容',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '发送时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    UNIQUE KEY uk_uid (uid),
    FOREIGN KEY (chat_room_uid) REFERENCES chat_rooms(uid),
    FOREIGN KEY (sender_uid) REFERENCES users(uid),
    INDEX idx_chat_room_time (chat_room_uid, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT '消息表';

-- Message media table
CREATE TABLE message_media (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT COMMENT '媒体记录ID',
    uid VARCHAR(36) NOT NULL DEFAULT (UUID()) COMMENT '媒体UUID',
    message_uid VARCHAR(36) NOT NULL COMMENT '消息UUID',
    media_type ENUM('image', 'file') NOT NULL COMMENT '媒体类型',
    media_url VARCHAR(255) NOT NULL COMMENT '媒体URL',
    file_name VARCHAR(255) COMMENT '文件名',
    file_size INT UNSIGNED COMMENT '文件大小',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    UNIQUE KEY uk_uid (uid),
    FOREIGN KEY (message_uid) REFERENCES messages(uid)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT '消息媒体表';

-- Pinned chat rooms table
CREATE TABLE pinned_chat_rooms (
    user_uid VARCHAR(36) NOT NULL COMMENT '用户UUID',
    chat_room_uid VARCHAR(36) NOT NULL COMMENT '聊天室UUID',
    pinned_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '置顶时间',
    PRIMARY KEY (user_uid, chat_room_uid),
    FOREIGN KEY (user_uid) REFERENCES users(uid),
    FOREIGN KEY (chat_room_uid) REFERENCES chat_rooms(uid)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT '聊天室置顶表';

-- Notification templates table
CREATE TABLE notification_templates (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT COMMENT '模板ID',
    uid VARCHAR(36) NOT NULL DEFAULT (UUID()) COMMENT '模板UUID',
    code VARCHAR(50) UNIQUE NOT NULL COMMENT '模板代码',
    title_template TEXT NOT NULL COMMENT '标题模板',
    content_template TEXT NOT NULL COMMENT '内容模板',
    platform ENUM('all', 'ios', 'android', 'web') DEFAULT 'all' COMMENT '平台类型',
    variables JSON COMMENT '模板变量',
    status ENUM('active', 'inactive') DEFAULT 'active' COMMENT '模板状态',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    UNIQUE KEY uk_uid (uid)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT '通知模板表';

-- Push notifications table
CREATE TABLE push_notifications (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT COMMENT '推送ID',
    uid VARCHAR(36) NOT NULL DEFAULT (UUID()) COMMENT '推送UUID',
    template_uid VARCHAR(36) COMMENT '模板UUID',
    title VARCHAR(255) NOT NULL COMMENT '推送标题',
    content TEXT NOT NULL COMMENT '推送内容',
    type ENUM('system', 'merchant', 'topic', 'chat') NOT NULL COMMENT '推送类型',
    sender_uid VARCHAR(36) COMMENT '发送者UUID',
    target_type ENUM('all', 'area', 'specific_users') NOT NULL COMMENT '目标类型',
    target_params JSON COMMENT '目标参数',
    status ENUM('draft', 'scheduled', 'sending', 'sent', 'cancelled') COMMENT '推送状态',
    scheduled_at TIMESTAMP COMMENT '计划发送时间',
    sent_at TIMESTAMP COMMENT '实际发送时间',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    UNIQUE KEY uk_uid (uid),
    FOREIGN KEY (template_uid) REFERENCES notification_templates(uid),
    FOREIGN KEY (sender_uid) REFERENCES users(uid)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT '推送消息表';

-- Push notification recipients table
CREATE TABLE push_notification_recipients (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT COMMENT '接收记录ID',
    uid VARCHAR(36) NOT NULL DEFAULT (UUID()) COMMENT '接收记录UUID',
    notification_uid VARCHAR(36) NOT NULL COMMENT '推送UUID',
    user_uid VARCHAR(36) NOT NULL COMMENT '接收者UUID',
    device_uid VARCHAR(36) NOT NULL COMMENT '接收设备UUID',
    status ENUM('pending', 'sent', 'failed', 'received', 'read') COMMENT '接收状态',
    error_message TEXT COMMENT '错误信息',
    sent_at TIMESTAMP COMMENT '发送时间',
    received_at TIMESTAMP COMMENT '接收时间',
    read_at TIMESTAMP COMMENT '阅读时间',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    UNIQUE KEY uk_uid (uid),
    UNIQUE KEY uk_notification_user (notification_uid, user_uid),  -- 保持原来的唯一约束
    FOREIGN KEY (notification_uid) REFERENCES push_notifications(uid),
    FOREIGN KEY (user_uid) REFERENCES users(uid),
    FOREIGN KEY (device_uid) REFERENCES user_devices(uid)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT '推送接收记录表';

-- Regions table
CREATE TABLE regions (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT COMMENT '地区ID',
    uid VARCHAR(36) NOT NULL DEFAULT (UUID()) COMMENT '地区UUID',
    parent_uid VARCHAR(36) COMMENT '父级地区UUID',
    name VARCHAR(50) NOT NULL COMMENT '地区名称',
    level ENUM('country', 'province', 'city', 'district') NOT NULL COMMENT '地区级别',
    code VARCHAR(20) UNIQUE COMMENT '地区编码',
    latitude DECIMAL(10, 8) COMMENT '中心纬度',
    longitude DECIMAL(11, 8) COMMENT '中心经度',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    UNIQUE KEY uk_uid (uid),
    FOREIGN KEY (parent_uid) REFERENCES regions(uid)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT '地区表';

-- User locations table
CREATE TABLE user_locations (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT COMMENT '记录ID',
    uid VARCHAR(36) NOT NULL DEFAULT (UUID()) COMMENT '位置记录UUID',
    user_uid VARCHAR(36) NOT NULL COMMENT '用户UUID',
    latitude DECIMAL(10, 8) NOT NULL COMMENT '纬度',
    longitude DECIMAL(11, 8) NOT NULL COMMENT '经度',
    region_uid VARCHAR(36) COMMENT '所属地区UUID',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '记录时间',
    UNIQUE KEY uk_uid (uid),
    FOREIGN KEY (user_uid) REFERENCES users(uid),
    FOREIGN KEY (region_uid) REFERENCES regions(uid),
    INDEX idx_user_location (user_uid, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT '用户位置历史表';

-- Sensitive words table
CREATE TABLE sensitive_words (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT COMMENT '敏感词ID',
    uid VARCHAR(36) NOT NULL DEFAULT (UUID()) COMMENT '敏感词UUID',
    word VARCHAR(50) NOT NULL COMMENT '敏感词',
    category VARCHAR(20) COMMENT '分类',
    level ENUM('low', 'medium', 'high') DEFAULT 'medium' COMMENT '级别',
    status BOOLEAN DEFAULT TRUE COMMENT '是否启用',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    UNIQUE KEY uk_uid (uid),
    UNIQUE KEY uk_word (word)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT '敏感词表';

-- Reports table
CREATE TABLE reports (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT COMMENT '举报ID',
    uid VARCHAR(36) NOT NULL DEFAULT (UUID()) COMMENT '举报UUID',
    reporter_uid VARCHAR(36) NOT NULL COMMENT '举报人UUID',
    target_type ENUM('user', 'topic', 'comment', 'message') NOT NULL COMMENT '举报对象类型',
    target_uid VARCHAR(36) NOT NULL COMMENT '举报对象UUID',
    reason_type VARCHAR(50) NOT NULL COMMENT '举报原因类型',
    reason_detail TEXT COMMENT '举报详细原因',
    status ENUM('pending', 'processing', 'resolved', 'rejected') DEFAULT 'pending' COMMENT '处理状态',
    handler_uid VARCHAR(36) COMMENT '处理人UUID',
    handling_result TEXT COMMENT '处理结果',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '举报时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    UNIQUE KEY uk_uid (uid),
    FOREIGN KEY (reporter_uid) REFERENCES users(uid),
    FOREIGN KEY (handler_uid) REFERENCES users(uid),
    INDEX idx_target (target_type, target_uid),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT '举报记录表';

-- Create additional indexes for range queries
CREATE INDEX idx_topic_location_time ON topics(location_latitude, location_longitude, created_at);
CREATE INDEX idx_user_active_location ON users(status, location_latitude, location_longitude);

-- Create full-text search indexes
ALTER TABLE topics ADD FULLTEXT INDEX ft_topic_content(title, content);

-- Create statistical indexes
CREATE INDEX idx_topic_stats ON topics(status, created_at);
CREATE INDEX idx_message_stats ON messages(created_at);

