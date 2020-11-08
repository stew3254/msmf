CREATE TABLE IF NOT EXISTS `game` (
  `id` SMALLINT PRIMARY KEY,
  `name` VARCHAR(64) NOT NULL
);

CREATE TABLE IF NOT EXISTS `version` (
  `id` VARCHAR(32) PRIMARY KEY,
  `game_id` SMALLINT NOT NULL FOREIGN KEY REFERENCES `game` (`id`)
);

CREATE TABLE IF NOT EXISTS `mod` (
  `id` INT PRIMARY KEY,
  `url` TEXT NOT NULL,
  `name` varchar(64) NOT NULL,
  `game_id` INT NOT NULL FOREIGN KEY REFERENCES `game` (`id`)
  `version_id` INT FOREIGN KEY REFERENCES `version` (`id`)
);

CREATE TABLE IF NOT EXISTS `server` (
  `id` SMALLINT PRIMARY KEY,
  `port` SMALLINT NOT NULL,
  `game_id` SMALLINT FOREIGN KEY REFERENCES `game` (`id`)
  `version_id` INT FOREIGN KEY REFERENCES `version` (`id`)
);