CREATE TABLE `imagespy_feature` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `created_at` datetime(6) NOT NULL,
  `name` varchar(255) NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE `imagespy_image` (
  `created_at` datetime(6) NOT NULL,
  `digest` varchar(71) NOT NULL,
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(255) NOT NULL,
  `schema_version` int(11) NOT NULL,
  `scraped_at` datetime(6) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `imagespy_image2_digest_name_0c92e169_uniq` (`digest`,`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE `imagespy_layer` (
  `digest` varchar(255) NOT NULL,
  `id` int(11) NOT NULL AUTO_INCREMENT,
  PRIMARY KEY (`id`),
  UNIQUE KEY `digest` (`digest`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE `imagespy_platform` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `architecture` varchar(255) NOT NULL,
  `created` datetime(6) DEFAULT NULL,
  `created_at` datetime(6) NOT NULL,
  `manifest_digest` varchar(255) NOT NULL,
  `os` varchar(255) NOT NULL,
  `os_version` varchar(255) DEFAULT NULL,
  `variant` varchar(255) DEFAULT NULL,
  `image_id` int(11) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `imagespy_platform2_image_id_manifest_digest_8c5aa328_uniq` (`image_id`,`manifest_digest`),
  CONSTRAINT `imagespy_platform_image_id_8c2f1dfe_fk_imagespy_image_id` FOREIGN KEY (`image_id`) REFERENCES `imagespy_image` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE `imagespy_tag` (
  `distinction` varchar(64) NOT NULL,
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `is_latest` tinyint(1) NOT NULL,
  `is_tagged` tinyint(1) NOT NULL,
  `name` varchar(50) NOT NULL,
  `image_id` int(11) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `imagespy_tag_distinction_image_id_name_4816e799_uniq` (`distinction`,`image_id`,`name`),
  KEY `imagespy_tag_image_id_c216b2ac_fk_imagespy_image2_id` (`image_id`),
  CONSTRAINT `imagespy_tag_image_id_c216b2ac_fk_imagespy_image2_id` FOREIGN KEY (`image_id`) REFERENCES `imagespy_image` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE `imagespy_layer_source_images` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `layer_id` int(11) NOT NULL,
  `image_id` int(11) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `imagespy_layer2_source_images_layer2_id_image2_id_fa7aef18_uniq` (`layer_id`,`image_id`),
  KEY `imagespy_layer_sourc_image_id_ce7add99_fk_imagespy_` (`image_id`),
  CONSTRAINT `imagespy_layer_sourc_image_id_ce7add99_fk_imagespy_` FOREIGN KEY (`image_id`) REFERENCES `imagespy_image` (`id`),
  CONSTRAINT `imagespy_layer_sourc_layer_id_f030ad8a_fk_imagespy_` FOREIGN KEY (`layer_id`) REFERENCES `imagespy_layer` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE `imagespy_layerofplatform` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `position` int(11) NOT NULL,
  `layer_id` int(11) NOT NULL,
  `platform_id` int(11) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `imagespy_layerofplatform2_layer_id_platform_id_c20c9139_uniq` (`layer_id`,`platform_id`),
  KEY `imagespy_layerofplat_platform_id_fedab095_fk_imagespy_` (`platform_id`),
  CONSTRAINT `imagespy_layerofplat_platform_id_fedab095_fk_imagespy_` FOREIGN KEY (`platform_id`) REFERENCES `imagespy_platform` (`id`),
  CONSTRAINT `imagespy_layerofplatform_layer_id_d57ed570_fk_imagespy_layer_id` FOREIGN KEY (`layer_id`) REFERENCES `imagespy_layer` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE `imagespy_osfeature` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `created_at` datetime(6) NOT NULL,
  `name` varchar(255) NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE `imagespy_platform_features` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `platform_id` int(11) NOT NULL,
  `feature_id` int(11) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `imagespy_platform2_featu_platform2_id_feature_id_e5cd780b_uniq` (`platform_id`,`feature_id`),
  KEY `imagespy_platform2_f_feature_id_60475c2e_fk_imagespy_` (`feature_id`),
  CONSTRAINT `imagespy_platform2_f_feature_id_60475c2e_fk_imagespy_` FOREIGN KEY (`feature_id`) REFERENCES `imagespy_feature` (`id`),
  CONSTRAINT `imagespy_platform_fe_platform_id_e10b625b_fk_imagespy_` FOREIGN KEY (`platform_id`) REFERENCES `imagespy_platform` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE `imagespy_platform_os_features` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `platform_id` int(11) NOT NULL,
  `osfeature_id` int(11) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `imagespy_platform2_os_fe_platform2_id_osfeature_i_d6f3d54f_uniq` (`platform_id`,`osfeature_id`),
  KEY `imagespy_platform2_o_osfeature_id_1f73791f_fk_imagespy_` (`osfeature_id`),
  CONSTRAINT `imagespy_platform2_o_osfeature_id_1f73791f_fk_imagespy_` FOREIGN KEY (`osfeature_id`) REFERENCES `imagespy_osfeature` (`id`),
  CONSTRAINT `imagespy_platform_os_platform_id_5ef73a47_fk_imagespy_` FOREIGN KEY (`platform_id`) REFERENCES `imagespy_platform` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
