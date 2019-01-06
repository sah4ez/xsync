CREATE DATABASE transaction_base
CHARACTER SET utf8
COLLATE utf8_general_ci;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
SET NAMES 'utf8';
USE transaction_base;
CREATE TABLE users (
  user_id int(11) UNSIGNED NOT NULL AUTO_INCREMENT,
  user_name varchar(50) DEFAULT NULL,
  user_nick varchar(255) DEFAULT NULL,
  PRIMARY KEY (user_id)
)
ENGINE = INNODB,
CHARACTER SET utf8,
COLLATE utf8_general_ci;
CREATE TABLE transaction_type (
  type_id smallint(6) NOT NULL AUTO_INCREMENT,
  type_name varchar(50) DEFAULT NULL,
  PRIMARY KEY (type_id)
)
ENGINE = INNODB,
CHARACTER SET utf8,
COLLATE utf8_general_ci;
CREATE TABLE transaction_status (
  status_id smallint(6) NOT NULL AUTO_INCREMENT,
  status_name varchar(50) DEFAULT NULL,
  PRIMARY KEY (status_id)
)
ENGINE = INNODB,
CHARACTER SET utf8,
COLLATE utf8_general_ci;
CREATE TABLE transactions (
  trx_id bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT,
  dt datetime DEFAULT NULL,
  type_id smallint(6) DEFAULT NULL,
  user_id int(11) UNSIGNED DEFAULT NULL,
  status_id smallint(6) DEFAULT NULL,
  PRIMARY KEY (trx_id)
)
ENGINE = INNODB,
CHARACTER SET utf8,
COLLATE utf8_general_ci;
ALTER TABLE transactions
ADD INDEX IDX_transactions_dt (dt);
ALTER TABLE transactions
ADD CONSTRAINT FK_transactions_status_id FOREIGN KEY (status_id)
REFERENCES transaction_status (status_id);
ALTER TABLE transactions
ADD CONSTRAINT FK_transactions_type_id FOREIGN KEY (type_id)
REFERENCES transaction_type (type_id);
ALTER TABLE transactions
ADD CONSTRAINT FK_transactions_user_id FOREIGN KEY (user_id)
REFERENCES users (user_id);
CREATE TABLE transaction_param_types (
  id_param_type smallint(6) UNSIGNED NOT NULL,
  name_param_type varchar(255) DEFAULT NULL,
  PRIMARY KEY (id_param_type)
)
ENGINE = INNODB,
CHARACTER SET utf8,
COLLATE utf8_general_ci;
CREATE TABLE transaction_params (
  id_tansaction bigint(20) UNSIGNED NOT NULL,
  id_param_type smallint(6) UNSIGNED NOT NULL,
  value varchar(255) DEFAULT NULL,
  PRIMARY KEY (id_tansaction, id_param_type)
)
ENGINE = INNODB,
CHARACTER SET utf8,
COLLATE utf8_general_ci;
ALTER TABLE transaction_params
ADD CONSTRAINT FK_transaction_params_id_param FOREIGN KEY (id_param_type)
REFERENCES transaction_param_types (id_param_type);
ALTER TABLE transaction_params
ADD CONSTRAINT FK_transaction_params_id_tansa FOREIGN KEY (id_tansaction)
REFERENCES transactions (trx_id) ON DELETE CASCADE ON UPDATE CASCADE;
/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS = @OLD_FOREIGN_KEY_CHECKS */;
