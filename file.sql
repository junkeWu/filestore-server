create table `tbl_file`
(
    `id`        int(11) not null auto_increment,
    `file_sha1` char(40)      not null default '' comment 'file_hash',
    `file_name` varchar(200)  not null default '' comment 'file_name',
    `file_size` bigint (20) not null default '0' comment 'file_size',
    `file_addr` varchar(1024) not null default '' comment 'file_store_address',
    `create_at` datetime default now() comment 'create_date',
    `update_at` datetime default now() on update current_timestamp () comment 'write_date',
    `status`    int(11) not null default '0' comment 'status(used/ban/stop)',
    `ext1`      int (11) default '0' comment 'spare field 1',
    `ext2`      text comment 'spare field 2',
    primary key (`id`),
    unique key `idx_file_hash`(`file_sha1`),
    key         `idx_status`(`status`)
)engine=InnoDB default charset=utf8;