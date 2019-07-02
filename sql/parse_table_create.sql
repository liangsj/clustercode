CREATE TABLE `praise_count` (
 `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT 'id',
 `resource_id` bigint(20) NOT NULL DEFAULT '0' COMMENT '资源id',
 `count` bigint(20) NOT NULL DEFAULT '0' COMMENT '次数',
 `ctime` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'create time',
 `mtime` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'modify time',
 PRIMARY KEY (`id`),
 UNIQUE KEY `resource_id` (`resource_id`)
) ENGINE=InnoDB AUTO_INCREMENT=1968180 DEFAULT CHARSET=latin1 COLLATE=latin1_bin ROW_FORMAT=COMPRESSED KEY_BLOCK_SIZE=8 COMMENT='praise_count'

CREATE USER 'test'@'%' IDENTIFIED BY 'test';
GRANT ALL ON *.* TO 'test'@'%';

ALTER USER 'test'@'%' IDENTIFIED WITH mysql_native_password BY 'test';
select user,plugin from mysql.user;


