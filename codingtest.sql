--
-- 表结构 `accounts`
--

CREATE TABLE `accounts` (
 `id` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL,
 `owner_id` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
 `balance` bigint(20) DEFAULT NULL,
 `status` tinyint(1) DEFAULT NULL COMMENT '1:normal 2:deleted',
 PRIMARY KEY (`id`),
 KEY `owner_id` (`owner_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;


-- --------------------------------------------------------

--
-- 表结构 `journals`
--

CREATE TABLE `journals` (
 `id` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL,
 `from_account_id` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
 `to_account_id` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
 `amount` bigint(20) DEFAULT NULL,
 `charge` smallint(20) DEFAULT NULL,
 `status` tinyint(1) DEFAULT NULL COMMENT '1:normal 2:processing 3:fail',
 `created_at` timestamp NULL DEFAULT NULL,
 PRIMARY KEY (`id`),
 KEY `from_account_id` (`from_account_id`),
 KEY `to_account_id` (`to_account_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;


-- --------------------------------------------------------

--
-- 表结构 `owners`
--

CREATE TABLE `owners` (
 `id` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL,
 `name` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
 PRIMARY KEY (`id`),
 KEY `name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

--
-- 插入两条测试数据 `owners`
--

INSERT INTO `owners` (`id`, `name`) VALUES
('bcmvipnbuiv3g12g827g', 'test1'),
('bcmvipnbuiv3g12g82ui', 'test2');