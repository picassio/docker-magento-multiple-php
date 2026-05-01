<?php
// Allow embedding in iframes (for Mage UI integration)
$cfg['AllowThirdPartyFraming'] = true;

// Auto-login with root (dev only)
$cfg['Servers'][1]['auth_type'] = 'config';
$cfg['Servers'][1]['user'] = 'root';
$cfg['Servers'][1]['password'] = 'root';

// Also connect to other DB services
$i = 2;
$cfg['Servers'][$i]['host'] = 'mysql80';
$cfg['Servers'][$i]['port'] = 3306;
$cfg['Servers'][$i]['auth_type'] = 'config';
$cfg['Servers'][$i]['user'] = 'root';
$cfg['Servers'][$i]['password'] = 'root';
$cfg['Servers'][$i]['verbose'] = 'MySQL 8.0';

$i = 3;
$cfg['Servers'][$i]['host'] = 'mariadb';
$cfg['Servers'][$i]['port'] = 3306;
$cfg['Servers'][$i]['auth_type'] = 'config';
$cfg['Servers'][$i]['user'] = 'root';
$cfg['Servers'][$i]['password'] = 'root';
$cfg['Servers'][$i]['verbose'] = 'MariaDB 11.4';
