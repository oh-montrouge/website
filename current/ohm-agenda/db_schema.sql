-- phpMyAdmin SQL Dump
-- version 5.2.2
-- https://www.phpmyadmin.net/
--
-- Hôte : ohmagendcyohm092.mysql.db
-- Généré le : ven. 10 avr. 2026 à 11:38
-- Version du serveur : 8.0.45-36
-- Version de PHP : 8.1.33

SET SQL_MODE = "NO_AUTO_VALUE_ON_ZERO";
START TRANSACTION;
SET time_zone = "+00:00";


/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8mb4 */;

--
-- Base de données : `ohmagendcyohm092`
--

-- --------------------------------------------------------

--
-- Structure de la table `concert`
--

CREATE TABLE `concert` (
  `ID_CONCERT` int NOT NULL,
  `NOM_CONCERT` varchar(200) NOT NULL,
  `DATE_CONCERT` date NOT NULL,
  `HEURE_RDV` time NOT NULL,
  `HEURE_CONCERT` time NOT NULL,
  `LIEU` varchar(100) NOT NULL,
  `ID_SAISON` int NOT NULL,
  `TENUE` varchar(100) DEFAULT NULL,
  `INFOS` varchar(200) DEFAULT NULL,
  `PROGRAMME` varchar(1000) DEFAULT NULL
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb3;

-- --------------------------------------------------------

--
-- Structure de la table `cotisation`
--

CREATE TABLE `cotisation` (
  `ID_COTISATION` int NOT NULL,
  `ID_SAISON` int NOT NULL,
  `ID_MUSICIEN` int NOT NULL,
  `PAIEMENT` tinyint(1) NOT NULL DEFAULT '0'
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb3;

-- --------------------------------------------------------

--
-- Structure de la table `evenement`
--

CREATE TABLE `evenement` (
  `ID_EVENT` int NOT NULL,
  `NOM_EVENT` varchar(500) NOT NULL,
  `DATE_EVENT` date NOT NULL,
  `HEURE` time NOT NULL,
  `LIEU` varchar(100) NOT NULL,
  `ID_SAISON` int NOT NULL,
  `INFOS` varchar(1000) CHARACTER SET utf8mb3 COLLATE utf8mb3_general_ci DEFAULT NULL,
  `ADRESSE` varchar(200) DEFAULT NULL,
  `TARIF` float DEFAULT NULL
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb3;

-- --------------------------------------------------------

--
-- Structure de la table `instrument`
--

CREATE TABLE `instrument` (
  `ID_INSTRUMENT` int NOT NULL,
  `NOM` varchar(50) NOT NULL
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb3;

-- --------------------------------------------------------

--
-- Structure de la table `musicien`
--

CREATE TABLE `musicien` (
  `ID_MUSICIEN` int NOT NULL,
  `PRENOM` varchar(50) NOT NULL,
  `NOM` varchar(50) NOT NULL,
  `ACTIF` tinyint(1) NOT NULL DEFAULT '1',
  `ID_INSTRUMENT` int NOT NULL
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb3;

-- --------------------------------------------------------

--
-- Structure de la table `presence_concert`
--

CREATE TABLE `presence_concert` (
  `ID_PRESENCE` int NOT NULL,
  `ID_CONCERT` int NOT NULL,
  `ID_MUSICIEN` int NOT NULL,
  `ID_INSTRUMENT` int NOT NULL,
  `STATUT` varchar(50) NOT NULL
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb3;

-- --------------------------------------------------------

--
-- Structure de la table `presence_event`
--

CREATE TABLE `presence_event` (
  `ID_PRESENCE_EVENT` int NOT NULL,
  `ID_EVENT` int NOT NULL,
  `ID_MUSICIEN` int NOT NULL,
  `NB_PRESENT` int NOT NULL,
  `NB_CHOIX_A1` int NOT NULL,
  `NB_CHOIX_A2` int NOT NULL,
  `NB_CHOIX_A3` int NOT NULL,
  `NB_CHOIX_B1` int NOT NULL,
  `NB_CHOIX_B2` int NOT NULL,
  `NB_CHOIX_B3` int NOT NULL,
  `NB_CHOIX_C1` int NOT NULL,
  `NB_CHOIX_C2` int NOT NULL,
  `NB_CHOIX_C3` int NOT NULL,
  `NB_CHOIX_D1` int NOT NULL,
  `NB_CHOIX_D2` int NOT NULL
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb3;

-- --------------------------------------------------------

--
-- Structure de la table `presence_repet`
--

CREATE TABLE `presence_repet` (
  `ID_PRESENCE` int NOT NULL,
  `ID_REPET` int NOT NULL,
  `ID_MUSICIEN` int NOT NULL,
  `STATUT` varchar(50) NOT NULL
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb3;

-- --------------------------------------------------------

--
-- Structure de la table `repetition`
--

CREATE TABLE `repetition` (
  `ID_REPET` int NOT NULL,
  `NOM_REPET` varchar(200) NOT NULL,
  `DATE_REPET` date NOT NULL,
  `HEURE_REPET` time NOT NULL,
  `NBRE_PLACE` int NOT NULL,
  `INFOS` varchar(1000) DEFAULT NULL,
  `PROGRAMME` varchar(1000) DEFAULT NULL
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb3;

-- --------------------------------------------------------

--
-- Structure de la table `saison`
--

CREATE TABLE `saison` (
  `ID_SAISON` int NOT NULL,
  `LIBELLE` varchar(50) NOT NULL
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb3;

-- --------------------------------------------------------

--
-- Structure de la table `statut_presence`
--

CREATE TABLE `statut_presence` (
  `ID_STATUT` int NOT NULL,
  `STATUT` varchar(50) NOT NULL,
  `IS_ADMIN` tinyint(1) NOT NULL
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb3;

-- --------------------------------------------------------

--
-- Structure de la table `wp_ahm_assets`
--

CREATE TABLE `wp_ahm_assets` (
  `ID` bigint NOT NULL,
  `path` text NOT NULL,
  `owner` int NOT NULL,
  `activities` text NOT NULL,
  `comments` text NOT NULL,
  `access` text NOT NULL,
  `metadata` text NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

-- --------------------------------------------------------

--
-- Structure de la table `wp_ahm_asset_links`
--

CREATE TABLE `wp_ahm_asset_links` (
  `ID` bigint NOT NULL,
  `asset_ID` bigint NOT NULL,
  `asset_key` varchar(255) NOT NULL,
  `access` text NOT NULL,
  `time` int NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

-- --------------------------------------------------------

--
-- Structure de la table `wp_ahm_download_stats`
--

CREATE TABLE `wp_ahm_download_stats` (
  `id` bigint NOT NULL,
  `pid` bigint NOT NULL,
  `uid` int NOT NULL,
  `oid` varchar(100) NOT NULL,
  `year` int NOT NULL,
  `month` int NOT NULL,
  `day` int NOT NULL,
  `timestamp` int NOT NULL,
  `ip` varchar(20) NOT NULL,
  `filename` text,
  `agent` text,
  `version` varchar(255) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

-- --------------------------------------------------------

--
-- Structure de la table `wp_ahm_emails`
--

CREATE TABLE `wp_ahm_emails` (
  `id` bigint NOT NULL,
  `email` varchar(255) NOT NULL,
  `pid` bigint NOT NULL,
  `date` int NOT NULL,
  `custom_data` text NOT NULL,
  `request_status` int NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

-- --------------------------------------------------------

--
-- Structure de la table `wp_ahm_sessions`
--

CREATE TABLE `wp_ahm_sessions` (
  `ID` bigint NOT NULL,
  `deviceID` varchar(255) NOT NULL,
  `name` varchar(255) NOT NULL,
  `value` text NOT NULL,
  `lastAccess` int NOT NULL,
  `expire` int NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

-- --------------------------------------------------------

--
-- Structure de la table `wp_ahm_social_conns`
--

CREATE TABLE `wp_ahm_social_conns` (
  `ID` bigint NOT NULL,
  `pid` bigint NOT NULL,
  `email` varchar(200) NOT NULL,
  `name` varchar(200) NOT NULL,
  `user_data` text NOT NULL,
  `access_token` text NOT NULL,
  `refresh_token` text NOT NULL,
  `source` varchar(200) NOT NULL,
  `timestamp` int NOT NULL,
  `processed` tinyint(1) NOT NULL DEFAULT '0'
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

-- --------------------------------------------------------

--
-- Structure de la table `wp_ahm_user_download_counts`
--

CREATE TABLE `wp_ahm_user_download_counts` (
  `ID` int NOT NULL,
  `user` varchar(255) NOT NULL,
  `package_id` int NOT NULL,
  `download_count` int NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

-- --------------------------------------------------------

--
-- Structure de la table `wp_commentmeta`
--

CREATE TABLE `wp_commentmeta` (
  `meta_id` bigint UNSIGNED NOT NULL,
  `comment_id` bigint UNSIGNED NOT NULL DEFAULT '0',
  `meta_key` varchar(255) COLLATE utf8mb4_unicode_520_ci DEFAULT NULL,
  `meta_value` longtext COLLATE utf8mb4_unicode_520_ci
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_520_ci;

-- --------------------------------------------------------

--
-- Structure de la table `wp_comments`
--

CREATE TABLE `wp_comments` (
  `comment_ID` bigint UNSIGNED NOT NULL,
  `comment_post_ID` bigint UNSIGNED NOT NULL DEFAULT '0',
  `comment_author` tinytext COLLATE utf8mb4_unicode_520_ci NOT NULL,
  `comment_author_email` varchar(100) COLLATE utf8mb4_unicode_520_ci NOT NULL DEFAULT '',
  `comment_author_url` varchar(200) COLLATE utf8mb4_unicode_520_ci NOT NULL DEFAULT '',
  `comment_author_IP` varchar(100) COLLATE utf8mb4_unicode_520_ci NOT NULL DEFAULT '',
  `comment_date` datetime NOT NULL DEFAULT '0000-00-00 00:00:00',
  `comment_date_gmt` datetime NOT NULL DEFAULT '0000-00-00 00:00:00',
  `comment_content` text COLLATE utf8mb4_unicode_520_ci NOT NULL,
  `comment_karma` int NOT NULL DEFAULT '0',
  `comment_approved` varchar(20) COLLATE utf8mb4_unicode_520_ci NOT NULL DEFAULT '1',
  `comment_agent` varchar(255) COLLATE utf8mb4_unicode_520_ci NOT NULL DEFAULT '',
  `comment_type` varchar(20) COLLATE utf8mb4_unicode_520_ci NOT NULL DEFAULT 'comment',
  `comment_parent` bigint UNSIGNED NOT NULL DEFAULT '0',
  `user_id` bigint UNSIGNED NOT NULL DEFAULT '0'
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_520_ci;

-- --------------------------------------------------------

--
-- Structure de la table `wp_links`
--

CREATE TABLE `wp_links` (
  `link_id` bigint UNSIGNED NOT NULL,
  `link_url` varchar(255) COLLATE utf8mb4_unicode_520_ci NOT NULL DEFAULT '',
  `link_name` varchar(255) COLLATE utf8mb4_unicode_520_ci NOT NULL DEFAULT '',
  `link_image` varchar(255) COLLATE utf8mb4_unicode_520_ci NOT NULL DEFAULT '',
  `link_target` varchar(25) COLLATE utf8mb4_unicode_520_ci NOT NULL DEFAULT '',
  `link_description` varchar(255) COLLATE utf8mb4_unicode_520_ci NOT NULL DEFAULT '',
  `link_visible` varchar(20) COLLATE utf8mb4_unicode_520_ci NOT NULL DEFAULT 'Y',
  `link_owner` bigint UNSIGNED NOT NULL DEFAULT '1',
  `link_rating` int NOT NULL DEFAULT '0',
  `link_updated` datetime NOT NULL DEFAULT '0000-00-00 00:00:00',
  `link_rel` varchar(255) COLLATE utf8mb4_unicode_520_ci NOT NULL DEFAULT '',
  `link_notes` mediumtext COLLATE utf8mb4_unicode_520_ci NOT NULL,
  `link_rss` varchar(255) COLLATE utf8mb4_unicode_520_ci NOT NULL DEFAULT ''
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_520_ci;

-- --------------------------------------------------------

--
-- Structure de la table `wp_options`
--

CREATE TABLE `wp_options` (
  `option_id` bigint UNSIGNED NOT NULL,
  `option_name` varchar(191) COLLATE utf8mb4_unicode_520_ci NOT NULL DEFAULT '',
  `option_value` longtext COLLATE utf8mb4_unicode_520_ci NOT NULL,
  `autoload` varchar(20) COLLATE utf8mb4_unicode_520_ci NOT NULL DEFAULT 'yes'
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_520_ci;

-- --------------------------------------------------------

--
-- Structure de la table `wp_postmeta`
--

CREATE TABLE `wp_postmeta` (
  `meta_id` bigint UNSIGNED NOT NULL,
  `post_id` bigint UNSIGNED NOT NULL DEFAULT '0',
  `meta_key` varchar(255) COLLATE utf8mb4_unicode_520_ci DEFAULT NULL,
  `meta_value` longtext COLLATE utf8mb4_unicode_520_ci
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_520_ci;

-- --------------------------------------------------------

--
-- Structure de la table `wp_posts`
--

CREATE TABLE `wp_posts` (
  `ID` bigint UNSIGNED NOT NULL,
  `post_author` bigint UNSIGNED NOT NULL DEFAULT '0',
  `post_date` datetime NOT NULL DEFAULT '0000-00-00 00:00:00',
  `post_date_gmt` datetime NOT NULL DEFAULT '0000-00-00 00:00:00',
  `post_content` longtext COLLATE utf8mb4_unicode_520_ci NOT NULL,
  `post_title` text COLLATE utf8mb4_unicode_520_ci NOT NULL,
  `post_excerpt` text COLLATE utf8mb4_unicode_520_ci NOT NULL,
  `post_status` varchar(20) COLLATE utf8mb4_unicode_520_ci NOT NULL DEFAULT 'publish',
  `comment_status` varchar(20) COLLATE utf8mb4_unicode_520_ci NOT NULL DEFAULT 'open',
  `ping_status` varchar(20) COLLATE utf8mb4_unicode_520_ci NOT NULL DEFAULT 'open',
  `post_password` varchar(255) COLLATE utf8mb4_unicode_520_ci NOT NULL DEFAULT '',
  `post_name` varchar(200) COLLATE utf8mb4_unicode_520_ci NOT NULL DEFAULT '',
  `to_ping` text COLLATE utf8mb4_unicode_520_ci NOT NULL,
  `pinged` text COLLATE utf8mb4_unicode_520_ci NOT NULL,
  `post_modified` datetime NOT NULL DEFAULT '0000-00-00 00:00:00',
  `post_modified_gmt` datetime NOT NULL DEFAULT '0000-00-00 00:00:00',
  `post_content_filtered` longtext COLLATE utf8mb4_unicode_520_ci NOT NULL,
  `post_parent` bigint UNSIGNED NOT NULL DEFAULT '0',
  `guid` varchar(255) COLLATE utf8mb4_unicode_520_ci NOT NULL DEFAULT '',
  `menu_order` int NOT NULL DEFAULT '0',
  `post_type` varchar(20) COLLATE utf8mb4_unicode_520_ci NOT NULL DEFAULT 'post',
  `post_mime_type` varchar(100) COLLATE utf8mb4_unicode_520_ci NOT NULL DEFAULT '',
  `comment_count` bigint NOT NULL DEFAULT '0'
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_520_ci;

-- --------------------------------------------------------

--
-- Structure de la table `wp_realmedialibrary`
--

CREATE TABLE `wp_realmedialibrary` (
  `id` mediumint NOT NULL,
  `parent` mediumint NOT NULL DEFAULT '-1',
  `name` tinytext COLLATE utf8mb4_unicode_520_ci NOT NULL,
  `slug` text COLLATE utf8mb4_unicode_520_ci NOT NULL,
  `absolute` text COLLATE utf8mb4_unicode_520_ci NOT NULL,
  `owner` bigint NOT NULL,
  `ord` mediumint NOT NULL DEFAULT '0',
  `oldCustomOrder` mediumint DEFAULT NULL,
  `contentCustomOrder` tinyint(1) NOT NULL DEFAULT '0',
  `type` varchar(10) COLLATE utf8mb4_unicode_520_ci NOT NULL DEFAULT '0',
  `restrictions` varchar(255) COLLATE utf8mb4_unicode_520_ci NOT NULL DEFAULT '',
  `cnt` mediumint DEFAULT NULL,
  `importId` bigint DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_520_ci;

-- --------------------------------------------------------

--
-- Structure de la table `wp_realmedialibrary_meta`
--

CREATE TABLE `wp_realmedialibrary_meta` (
  `meta_id` bigint UNSIGNED NOT NULL,
  `realmedialibrary_id` bigint UNSIGNED NOT NULL DEFAULT '0',
  `meta_key` varchar(255) COLLATE utf8mb4_unicode_520_ci DEFAULT NULL,
  `meta_value` longtext COLLATE utf8mb4_unicode_520_ci
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_520_ci;

-- --------------------------------------------------------

--
-- Structure de la table `wp_realmedialibrary_posts`
--

CREATE TABLE `wp_realmedialibrary_posts` (
  `attachment` bigint NOT NULL,
  `fid` mediumint NOT NULL DEFAULT '-1',
  `isShortcut` bigint NOT NULL DEFAULT '0',
  `nr` bigint DEFAULT NULL,
  `oldCustomNr` bigint DEFAULT NULL,
  `importData` text COLLATE utf8mb4_unicode_520_ci
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_520_ci;

-- --------------------------------------------------------

--
-- Structure de la table `wp_realmedialibrary_tmp`
--

CREATE TABLE `wp_realmedialibrary_tmp` (
  `id` mediumint NOT NULL,
  `parent` mediumint NOT NULL DEFAULT '-1',
  `name` tinytext COLLATE utf8mb4_unicode_520_ci NOT NULL,
  `ord` mediumint NOT NULL DEFAULT '0'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_520_ci;

-- --------------------------------------------------------

--
-- Structure de la table `wp_termmeta`
--

CREATE TABLE `wp_termmeta` (
  `meta_id` bigint UNSIGNED NOT NULL,
  `term_id` bigint UNSIGNED NOT NULL DEFAULT '0',
  `meta_key` varchar(255) COLLATE utf8mb4_unicode_520_ci DEFAULT NULL,
  `meta_value` longtext COLLATE utf8mb4_unicode_520_ci
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_520_ci;

-- --------------------------------------------------------

--
-- Structure de la table `wp_terms`
--

CREATE TABLE `wp_terms` (
  `term_id` bigint UNSIGNED NOT NULL,
  `name` varchar(200) COLLATE utf8mb4_unicode_520_ci NOT NULL DEFAULT '',
  `slug` varchar(200) COLLATE utf8mb4_unicode_520_ci NOT NULL DEFAULT '',
  `term_group` bigint NOT NULL DEFAULT '0'
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_520_ci;

-- --------------------------------------------------------

--
-- Structure de la table `wp_term_relationships`
--

CREATE TABLE `wp_term_relationships` (
  `object_id` bigint UNSIGNED NOT NULL DEFAULT '0',
  `term_taxonomy_id` bigint UNSIGNED NOT NULL DEFAULT '0',
  `term_order` int NOT NULL DEFAULT '0'
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_520_ci;

-- --------------------------------------------------------

--
-- Structure de la table `wp_term_taxonomy`
--

CREATE TABLE `wp_term_taxonomy` (
  `term_taxonomy_id` bigint UNSIGNED NOT NULL,
  `term_id` bigint UNSIGNED NOT NULL DEFAULT '0',
  `taxonomy` varchar(32) COLLATE utf8mb4_unicode_520_ci NOT NULL DEFAULT '',
  `description` longtext COLLATE utf8mb4_unicode_520_ci NOT NULL,
  `parent` bigint UNSIGNED NOT NULL DEFAULT '0',
  `count` bigint NOT NULL DEFAULT '0'
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_520_ci;

-- --------------------------------------------------------

--
-- Structure de la table `wp_um_metadata`
--

CREATE TABLE `wp_um_metadata` (
  `umeta_id` bigint UNSIGNED NOT NULL,
  `user_id` bigint UNSIGNED NOT NULL DEFAULT '0',
  `um_key` varchar(255) COLLATE utf8mb4_unicode_520_ci DEFAULT NULL,
  `um_value` longtext COLLATE utf8mb4_unicode_520_ci
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_520_ci;

-- --------------------------------------------------------

--
-- Structure de la table `wp_usermeta`
--

CREATE TABLE `wp_usermeta` (
  `umeta_id` bigint UNSIGNED NOT NULL,
  `user_id` bigint UNSIGNED NOT NULL DEFAULT '0',
  `meta_key` varchar(255) COLLATE utf8mb4_unicode_520_ci DEFAULT NULL,
  `meta_value` longtext COLLATE utf8mb4_unicode_520_ci
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_520_ci;

-- --------------------------------------------------------

--
-- Structure de la table `wp_users`
--

CREATE TABLE `wp_users` (
  `ID` bigint UNSIGNED NOT NULL,
  `user_login` varchar(60) COLLATE utf8mb4_unicode_520_ci NOT NULL DEFAULT '',
  `user_pass` varchar(255) COLLATE utf8mb4_unicode_520_ci NOT NULL DEFAULT '',
  `user_nicename` varchar(50) COLLATE utf8mb4_unicode_520_ci NOT NULL DEFAULT '',
  `user_email` varchar(100) COLLATE utf8mb4_unicode_520_ci NOT NULL DEFAULT '',
  `user_url` varchar(100) COLLATE utf8mb4_unicode_520_ci NOT NULL DEFAULT '',
  `user_registered` datetime NOT NULL DEFAULT '0000-00-00 00:00:00',
  `user_activation_key` varchar(255) COLLATE utf8mb4_unicode_520_ci NOT NULL DEFAULT '',
  `user_status` int NOT NULL DEFAULT '0',
  `display_name` varchar(250) COLLATE utf8mb4_unicode_520_ci NOT NULL DEFAULT ''
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_520_ci;

-- --------------------------------------------------------

--
-- Structure de la table `wp_wpfm_backup`
--

CREATE TABLE `wp_wpfm_backup` (
  `id` int NOT NULL,
  `backup_name` text COLLATE utf8mb4_unicode_520_ci,
  `backup_date` text COLLATE utf8mb4_unicode_520_ci
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_520_ci;

--
-- Index pour les tables déchargées
--

--
-- Index pour la table `concert`
--
ALTER TABLE `concert`
  ADD PRIMARY KEY (`ID_CONCERT`);

--
-- Index pour la table `cotisation`
--
ALTER TABLE `cotisation`
  ADD PRIMARY KEY (`ID_COTISATION`);

--
-- Index pour la table `evenement`
--
ALTER TABLE `evenement`
  ADD PRIMARY KEY (`ID_EVENT`);

--
-- Index pour la table `instrument`
--
ALTER TABLE `instrument`
  ADD PRIMARY KEY (`ID_INSTRUMENT`);

--
-- Index pour la table `musicien`
--
ALTER TABLE `musicien`
  ADD PRIMARY KEY (`ID_MUSICIEN`);

--
-- Index pour la table `presence_concert`
--
ALTER TABLE `presence_concert`
  ADD PRIMARY KEY (`ID_PRESENCE`);

--
-- Index pour la table `presence_event`
--
ALTER TABLE `presence_event`
  ADD PRIMARY KEY (`ID_PRESENCE_EVENT`);

--
-- Index pour la table `presence_repet`
--
ALTER TABLE `presence_repet`
  ADD PRIMARY KEY (`ID_PRESENCE`);

--
-- Index pour la table `repetition`
--
ALTER TABLE `repetition`
  ADD PRIMARY KEY (`ID_REPET`);

--
-- Index pour la table `saison`
--
ALTER TABLE `saison`
  ADD PRIMARY KEY (`ID_SAISON`);

--
-- Index pour la table `statut_presence`
--
ALTER TABLE `statut_presence`
  ADD PRIMARY KEY (`ID_STATUT`);

--
-- Index pour la table `wp_ahm_assets`
--
ALTER TABLE `wp_ahm_assets`
  ADD PRIMARY KEY (`ID`);

--
-- Index pour la table `wp_ahm_asset_links`
--
ALTER TABLE `wp_ahm_asset_links`
  ADD PRIMARY KEY (`ID`),
  ADD UNIQUE KEY `asset_key` (`asset_key`);

--
-- Index pour la table `wp_ahm_download_stats`
--
ALTER TABLE `wp_ahm_download_stats`
  ADD PRIMARY KEY (`id`);

--
-- Index pour la table `wp_ahm_emails`
--
ALTER TABLE `wp_ahm_emails`
  ADD PRIMARY KEY (`id`);

--
-- Index pour la table `wp_ahm_sessions`
--
ALTER TABLE `wp_ahm_sessions`
  ADD PRIMARY KEY (`ID`);

--
-- Index pour la table `wp_ahm_social_conns`
--
ALTER TABLE `wp_ahm_social_conns`
  ADD PRIMARY KEY (`ID`);

--
-- Index pour la table `wp_ahm_user_download_counts`
--
ALTER TABLE `wp_ahm_user_download_counts`
  ADD PRIMARY KEY (`ID`);

--
-- Index pour la table `wp_commentmeta`
--
ALTER TABLE `wp_commentmeta`
  ADD PRIMARY KEY (`meta_id`),
  ADD KEY `comment_id` (`comment_id`),
  ADD KEY `meta_key` (`meta_key`(191));

--
-- Index pour la table `wp_comments`
--
ALTER TABLE `wp_comments`
  ADD PRIMARY KEY (`comment_ID`),
  ADD KEY `comment_post_ID` (`comment_post_ID`),
  ADD KEY `comment_approved_date_gmt` (`comment_approved`,`comment_date_gmt`),
  ADD KEY `comment_date_gmt` (`comment_date_gmt`),
  ADD KEY `comment_parent` (`comment_parent`),
  ADD KEY `comment_author_email` (`comment_author_email`(10));

--
-- Index pour la table `wp_links`
--
ALTER TABLE `wp_links`
  ADD PRIMARY KEY (`link_id`),
  ADD KEY `link_visible` (`link_visible`);

--
-- Index pour la table `wp_options`
--
ALTER TABLE `wp_options`
  ADD PRIMARY KEY (`option_id`),
  ADD UNIQUE KEY `option_name` (`option_name`),
  ADD KEY `autoload` (`autoload`);

--
-- Index pour la table `wp_postmeta`
--
ALTER TABLE `wp_postmeta`
  ADD PRIMARY KEY (`meta_id`),
  ADD KEY `post_id` (`post_id`),
  ADD KEY `meta_key` (`meta_key`(191));

--
-- Index pour la table `wp_posts`
--
ALTER TABLE `wp_posts`
  ADD PRIMARY KEY (`ID`),
  ADD KEY `post_name` (`post_name`(191)),
  ADD KEY `type_status_date` (`post_type`,`post_status`,`post_date`,`ID`),
  ADD KEY `post_parent` (`post_parent`),
  ADD KEY `post_author` (`post_author`);

--
-- Index pour la table `wp_realmedialibrary`
--
ALTER TABLE `wp_realmedialibrary`
  ADD PRIMARY KEY (`id`);

--
-- Index pour la table `wp_realmedialibrary_meta`
--
ALTER TABLE `wp_realmedialibrary_meta`
  ADD PRIMARY KEY (`meta_id`),
  ADD KEY `realmedialibrary_id` (`realmedialibrary_id`),
  ADD KEY `meta_key` (`meta_key`(191));

--
-- Index pour la table `wp_realmedialibrary_posts`
--
ALTER TABLE `wp_realmedialibrary_posts`
  ADD PRIMARY KEY (`attachment`,`isShortcut`),
  ADD KEY `rmljoin` (`attachment`,`fid`);

--
-- Index pour la table `wp_realmedialibrary_tmp`
--
ALTER TABLE `wp_realmedialibrary_tmp`
  ADD PRIMARY KEY (`id`);

--
-- Index pour la table `wp_termmeta`
--
ALTER TABLE `wp_termmeta`
  ADD PRIMARY KEY (`meta_id`),
  ADD KEY `term_id` (`term_id`),
  ADD KEY `meta_key` (`meta_key`(191));

--
-- Index pour la table `wp_terms`
--
ALTER TABLE `wp_terms`
  ADD PRIMARY KEY (`term_id`),
  ADD KEY `slug` (`slug`(191)),
  ADD KEY `name` (`name`(191));

--
-- Index pour la table `wp_term_relationships`
--
ALTER TABLE `wp_term_relationships`
  ADD PRIMARY KEY (`object_id`,`term_taxonomy_id`),
  ADD KEY `term_taxonomy_id` (`term_taxonomy_id`);

--
-- Index pour la table `wp_term_taxonomy`
--
ALTER TABLE `wp_term_taxonomy`
  ADD PRIMARY KEY (`term_taxonomy_id`),
  ADD UNIQUE KEY `term_id_taxonomy` (`term_id`,`taxonomy`),
  ADD KEY `taxonomy` (`taxonomy`);

--
-- Index pour la table `wp_um_metadata`
--
ALTER TABLE `wp_um_metadata`
  ADD PRIMARY KEY (`umeta_id`),
  ADD KEY `user_id_indx` (`user_id`),
  ADD KEY `meta_key_indx` (`um_key`(191)),
  ADD KEY `meta_value_indx` (`um_value`(191));

--
-- Index pour la table `wp_usermeta`
--
ALTER TABLE `wp_usermeta`
  ADD PRIMARY KEY (`umeta_id`),
  ADD KEY `user_id` (`user_id`),
  ADD KEY `meta_key` (`meta_key`(191));

--
-- Index pour la table `wp_users`
--
ALTER TABLE `wp_users`
  ADD PRIMARY KEY (`ID`),
  ADD KEY `user_login_key` (`user_login`),
  ADD KEY `user_nicename` (`user_nicename`),
  ADD KEY `user_email` (`user_email`);

--
-- Index pour la table `wp_wpfm_backup`
--
ALTER TABLE `wp_wpfm_backup`
  ADD PRIMARY KEY (`id`);

--
-- AUTO_INCREMENT pour les tables déchargées
--

--
-- AUTO_INCREMENT pour la table `concert`
--
ALTER TABLE `concert`
  MODIFY `ID_CONCERT` int NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT pour la table `cotisation`
--
ALTER TABLE `cotisation`
  MODIFY `ID_COTISATION` int NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT pour la table `evenement`
--
ALTER TABLE `evenement`
  MODIFY `ID_EVENT` int NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT pour la table `instrument`
--
ALTER TABLE `instrument`
  MODIFY `ID_INSTRUMENT` int NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT pour la table `musicien`
--
ALTER TABLE `musicien`
  MODIFY `ID_MUSICIEN` int NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT pour la table `presence_concert`
--
ALTER TABLE `presence_concert`
  MODIFY `ID_PRESENCE` int NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT pour la table `presence_event`
--
ALTER TABLE `presence_event`
  MODIFY `ID_PRESENCE_EVENT` int NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT pour la table `presence_repet`
--
ALTER TABLE `presence_repet`
  MODIFY `ID_PRESENCE` int NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT pour la table `repetition`
--
ALTER TABLE `repetition`
  MODIFY `ID_REPET` int NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT pour la table `saison`
--
ALTER TABLE `saison`
  MODIFY `ID_SAISON` int NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT pour la table `statut_presence`
--
ALTER TABLE `statut_presence`
  MODIFY `ID_STATUT` int NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT pour la table `wp_ahm_assets`
--
ALTER TABLE `wp_ahm_assets`
  MODIFY `ID` bigint NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT pour la table `wp_ahm_asset_links`
--
ALTER TABLE `wp_ahm_asset_links`
  MODIFY `ID` bigint NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT pour la table `wp_ahm_download_stats`
--
ALTER TABLE `wp_ahm_download_stats`
  MODIFY `id` bigint NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT pour la table `wp_ahm_emails`
--
ALTER TABLE `wp_ahm_emails`
  MODIFY `id` bigint NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT pour la table `wp_ahm_sessions`
--
ALTER TABLE `wp_ahm_sessions`
  MODIFY `ID` bigint NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT pour la table `wp_ahm_social_conns`
--
ALTER TABLE `wp_ahm_social_conns`
  MODIFY `ID` bigint NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT pour la table `wp_ahm_user_download_counts`
--
ALTER TABLE `wp_ahm_user_download_counts`
  MODIFY `ID` int NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT pour la table `wp_commentmeta`
--
ALTER TABLE `wp_commentmeta`
  MODIFY `meta_id` bigint UNSIGNED NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT pour la table `wp_comments`
--
ALTER TABLE `wp_comments`
  MODIFY `comment_ID` bigint UNSIGNED NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT pour la table `wp_links`
--
ALTER TABLE `wp_links`
  MODIFY `link_id` bigint UNSIGNED NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT pour la table `wp_options`
--
ALTER TABLE `wp_options`
  MODIFY `option_id` bigint UNSIGNED NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT pour la table `wp_postmeta`
--
ALTER TABLE `wp_postmeta`
  MODIFY `meta_id` bigint UNSIGNED NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT pour la table `wp_posts`
--
ALTER TABLE `wp_posts`
  MODIFY `ID` bigint UNSIGNED NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT pour la table `wp_realmedialibrary`
--
ALTER TABLE `wp_realmedialibrary`
  MODIFY `id` mediumint NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT pour la table `wp_realmedialibrary_meta`
--
ALTER TABLE `wp_realmedialibrary_meta`
  MODIFY `meta_id` bigint UNSIGNED NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT pour la table `wp_realmedialibrary_tmp`
--
ALTER TABLE `wp_realmedialibrary_tmp`
  MODIFY `id` mediumint NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT pour la table `wp_termmeta`
--
ALTER TABLE `wp_termmeta`
  MODIFY `meta_id` bigint UNSIGNED NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT pour la table `wp_terms`
--
ALTER TABLE `wp_terms`
  MODIFY `term_id` bigint UNSIGNED NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT pour la table `wp_term_taxonomy`
--
ALTER TABLE `wp_term_taxonomy`
  MODIFY `term_taxonomy_id` bigint UNSIGNED NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT pour la table `wp_um_metadata`
--
ALTER TABLE `wp_um_metadata`
  MODIFY `umeta_id` bigint UNSIGNED NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT pour la table `wp_usermeta`
--
ALTER TABLE `wp_usermeta`
  MODIFY `umeta_id` bigint UNSIGNED NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT pour la table `wp_users`
--
ALTER TABLE `wp_users`
  MODIFY `ID` bigint UNSIGNED NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT pour la table `wp_wpfm_backup`
--
ALTER TABLE `wp_wpfm_backup`
  MODIFY `id` int NOT NULL AUTO_INCREMENT;
COMMIT;

/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
