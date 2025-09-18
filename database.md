# MySQL SQL

## 重要说明

### Transactions 表
- `signature` 字段：交易的唯一签名标识符，用于数据库存储和内部引用
- `index_number` 字段：交易的索引号，**用作 REST API 的 URL 参数**
- API 路由格式：`/multisigs/{multisig_address}/transactions/{indexNumber}`

```
-- ---------- Multisigs ----------
CREATE TABLE IF NOT EXISTS `multisigs` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `multisig_address` VARCHAR(255) NOT NULL,
  `name` VARCHAR(255),
  `description` TEXT,
  `created_at` DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_multisigs_multisig_address` (`multisig_address`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ---------- Vaults ----------
CREATE TABLE IF NOT EXISTS `vaults` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `vault_address` VARCHAR(255) NOT NULL,
  `multisig_address` VARCHAR(255) NOT NULL,
  `name` VARCHAR(255),
  `description` TEXT,
  `created_at` DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_vaults_vault_address` (`vault_address`),
  KEY `idx_vaults_multisig_address` (`multisig_address`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ---------- Members ----------
CREATE TABLE IF NOT EXISTS `members` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `member_address` VARCHAR(255) NOT NULL,
  `multisig_address` VARCHAR(255) NOT NULL,
  `name` VARCHAR(255),
  `description` TEXT,
  `created_at` DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_members_member_address` (`member_address`),
  KEY `idx_members_multisig_address` (`multisig_address`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ---------- Spends ----------
CREATE TABLE IF NOT EXISTS `spends` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `spend_address` VARCHAR(255) NOT NULL,
  `multisig_address` VARCHAR(255) NOT NULL,
  `name` VARCHAR(255),
  `description` TEXT,
  `created_at` DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_spends_spend_address` (`spend_address`),
  KEY `idx_spends_multisig_address` (`multisig_address`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ---------- Transactions ----------
CREATE TABLE IF NOT EXISTS `transactions` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `signature` VARCHAR(255) NOT NULL,
  `multisig_address` VARCHAR(255) NOT NULL,
  `index_number` BIGINT NOT NULL,
  `created_at` DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_transactions_signature` (`signature`),
  KEY `idx_transactions_multisig_address` (`multisig_address`),
  KEY `idx_transactions_index_number` (`index_number`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

```
