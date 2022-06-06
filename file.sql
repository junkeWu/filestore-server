create table `tbl_file`
(
    `id`        int(11)       not null auto_increment,
    `file_sha1` char(40)      not null default '' comment 'file_hash',
    `file_name` varchar(200)  not null default '' comment 'file_name',
    `file_size` bigint(20)    not null default '0' comment 'file_size',
    `file_addr` varchar(1024) not null default '' comment 'file_store_address',
    `create_at` datetime               default now() comment 'create_date',
    `update_at` datetime               default now() on update current_timestamp() comment 'write_date',
    `status`    int(11)       not null default '0' comment 'status(used/ban/stop)',
    `ext1`      int(11)                default '0' comment 'spare field 1',
    `ext2`      text comment 'spare field 2',
    primary key (`id`),
    unique key `idx_file_hash` (`file_sha1`),
    key `idx_status` (`status`)
) engine = InnoDB
  default charset = utf8;


# 用户表
CREATE TABLE tbl_user
(
    `id`              int(11)      NOT NULL AUTO_INCREMENT,
    `user_name`       varchar(64)  NOT NULL DEFAULT '' COMMENT '用户名',
    `user_pwd`        varchar(256) NOT NULL DEFAULT '' COMMENT '用户encoded密码',
    `email`           varchar(64)           DEFAULT '' COMMENT '邮箱',
    `phone`           varchar(128)          DEFAULT '' COMMENT '手机号',
    `email_validated` tinyint(1)            DEFAULT 0 COMMENT '邮箱是否已验证',
    `phone_validated` tinyint(11)           DEFAULT 0 COMMENT '手机号是否已验证',
    `signup_at`       datetime              DEFAULT CURRENT_TIMESTAMP COMMENT '注册日期',
    `last_active`     datetime              DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后话跃时间戳',
    `profile`         text COMMENT '用户属性',
    `status`          int(11)      NOT NULL DEFAULT 0 COMMENT '账户状态（启用/禁用/锁定/标记删除等)',
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_phone` (`phone`),
    KEY `idx_status` (`status`)
) ENGINE = InnoDB
  AUTO_INCREMENT = 5
  DEFAULT CHARSET = utf8mb4;


# token
CREATE TABLE tbl_user_token
(
    id         int(11)     NOT NULL AUTO_INCREMENT,
    user_name  varchar(64) NOT NULL DEFAULT '' COMMENT '用户名',
    user_token char(40)    NOT NULL DEFAULT '' COMMENT '用户登录token',
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_username` (`user_name`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4;


# 用户文件表
CREATE TABLE tbl_user_file
(
    `id` int(11) NOT NULL PRIMARY KEY AUTO_INCREMENT,
    `user_name` varchar(64) NOT NULL,
    `file_sha1` varchar(64) NOT NULL DEFAULT '' COMMENT 'hash',
    `file_size` bigint(20) DEFAULT 0 COMMENT '文件大小',
    `file_name` varchar(256) NOT NULL DEFAULT '' COMMENT '文件名',
    `upload_at`  datetime     default CURRENT_TIMESTAMP  comment '上传时间',
    last_update  datetime     default CURRENT_TIMESTAMP  on update CURRENT_TIMESTAMP comment '最后修改时间',
    `status` int(11) NOT NULL DEFAULT '0' COMMENT '文件状态(O正常1已删除2禁用)',
    UNIQUE KEY  `idx_user_file` (`user_name`,`file_sha1`),
    KEY `idx_status` (`status`),
    KEY `idx_user_id` (`user_name`)
) ENGINE=InnoDB
  DEFAULT
  CHARSET=utf8mb4;