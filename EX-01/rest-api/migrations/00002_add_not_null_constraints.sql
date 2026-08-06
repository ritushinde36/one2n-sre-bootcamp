-- +goose Up
ALTER TABLE `students`
  MODIFY COLUMN `name` varchar(255) NOT NULL,
  MODIFY COLUMN `email` varchar(191) NOT NULL,
  MODIFY COLUMN `age` bigint NOT NULL,
  MODIFY COLUMN `class` varchar(255) NOT NULL,
  MODIFY COLUMN `department` varchar(255) NOT NULL;

-- +goose Down
ALTER TABLE `students`
  MODIFY COLUMN `name` longtext,
  MODIFY COLUMN `email` varchar(191) DEFAULT NULL,
  MODIFY COLUMN `age` bigint DEFAULT NULL,
  MODIFY COLUMN `class` longtext,
  MODIFY COLUMN `department` longtext;
