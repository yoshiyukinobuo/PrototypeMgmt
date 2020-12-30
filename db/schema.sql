SET NAMES utf8mb4;

CREATE TABLE prototype_requests (
  id           BIGINT AUTO_INCREMENT PRIMARY KEY,
  title        VARCHAR(200) NOT NULL,
  product_type ENUM('GPS','ANTENNA','OTHER') NOT NULL DEFAULT 'OTHER',
  customer     VARCHAR(200) NULL,
  quantity     INT NOT NULL DEFAULT 1,
  due_date     DATE NULL,
  status       ENUM('REQUESTED','APPROVED','IN_BUILD','EVALUATING','DONE','REJECTED') NOT NULL DEFAULT 'REQUESTED',
  created_at   TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB;

CREATE TABLE erp_items (
  item_code  VARCHAR(140) PRIMARY KEY,
  item_name  VARCHAR(255) NOT NULL,
  item_group VARCHAR(140) NULL,
  synced_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB;

CREATE TABLE prototype_lots (
  id             BIGINT AUTO_INCREMENT PRIMARY KEY,
  lot_no         VARCHAR(40) NOT NULL UNIQUE,
  request_id     BIGINT NULL,
  item_code      VARCHAR(140) NULL,
  bom_revision   VARCHAR(20) NOT NULL DEFAULT '',
  quantity       INT NOT NULL,
  stage          ENUM('PLANNED','BUILD','INSPECTION','SHIPPED','EVALUATION','CLOSED') NOT NULL DEFAULT 'PLANNED',
  erp_work_order VARCHAR(140) NULL,
  created_at     TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (request_id) REFERENCES prototype_requests(id),
  FOREIGN KEY (item_code)  REFERENCES erp_items(item_code)
) ENGINE=InnoDB;

CREATE TABLE units (
  id         BIGINT AUTO_INCREMENT PRIMARY KEY,
  serial_no  VARCHAR(60) NOT NULL UNIQUE,
  lot_id     BIGINT NOT NULL,
  status     ENUM('BUILT','PASS','FAIL','SCRAPPED','SHIPPED') NOT NULL DEFAULT 'BUILT',
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (lot_id) REFERENCES prototype_lots(id)
) ENGINE=InnoDB;

CREATE TABLE test_results (
  id        BIGINT AUTO_INCREMENT PRIMARY KEY,
  unit_id   BIGINT NOT NULL,
  test_name VARCHAR(100) NOT NULL,
  value     DOUBLE NULL,
  unit      VARCHAR(20) NOT NULL DEFAULT '',
  spec_min  DOUBLE NULL,
  spec_max  DOUBLE NULL,
  passed    BOOLEAN NOT NULL,
  file_url  VARCHAR(500) NOT NULL DEFAULT '',
  tested_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (unit_id) REFERENCES units(id)
) ENGINE=InnoDB;

CREATE TABLE issues (
  id             BIGINT AUTO_INCREMENT PRIMARY KEY,
  lot_id         BIGINT NULL,
  title          VARCHAR(200) NOT NULL,
  description    TEXT NULL,
  severity       ENUM('LOW','MEDIUM','HIGH') NOT NULL DEFAULT 'MEDIUM',
  status         ENUM('OPEN','ANALYZING','FIXED','CLOSED') NOT NULL DEFAULT 'OPEN',
  countermeasure TEXT NULL,
  created_at     TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (lot_id) REFERENCES prototype_lots(id)
) ENGINE=InnoDB;

CREATE TABLE erp_sync_log (
  id         BIGINT AUTO_INCREMENT PRIMARY KEY,
  direction  ENUM('IN','OUT') NOT NULL,
  entity     VARCHAR(50) NOT NULL,
  status     ENUM('OK','ERROR') NOT NULL,
  message    TEXT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB;

-- メモ 017: テーブル間の外部キーと状態値を確認。状態値の表記を統一。
